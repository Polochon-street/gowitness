package writers

import (
	"sync"

	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/database"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"gorm.io/gorm"
)

var hammingThreshold = 10

// DbWriter is a Database writer
type DbWriter struct {
	URI           string
	conn          *gorm.DB
	mutex         sync.Mutex
	hammingGroups []islazy.HammingGroup
}

// NewDbWriter initialises a database writer
func NewDbWriter(uri string, debug bool) (*DbWriter, error) {
	c, err := database.Connection(uri, false, debug)
	if err != nil {
		return nil, err
	}

	return &DbWriter{
		URI:           uri,
		conn:          c,
		mutex:         sync.Mutex{},
		hammingGroups: []islazy.HammingGroup{},
	}, nil
}

// Write results to the database
func (dw *DbWriter) Write(result *models.Result) error {
	dw.mutex.Lock()
	defer dw.mutex.Unlock()

	// Assign Group ID based on PerceptionHash
	groupID, err := dw.AssignGroupID(result.PerceptionHash)
	if err == nil {
		result.PerceptionHashGroupId = groupID
	} else {
		// if we couldn't get a perception hash, thats okay. maybe the
		// screenshot failed.
		log.Debug("could not get group id for perception hash", "hash", result.PerceptionHash)
	}

	// Extract associations before creating the result
	cookies := result.Cookies
	headers := result.Headers
	networkLogs := result.Network
	consoleLogs := result.Console
	technologies := result.Technologies
	tls := result.TLS
	tags := result.Tags

	// Clear associations to insert the main result first
	result.Cookies = nil
	result.Headers = nil
	result.Network = nil
	result.Console = nil
	result.Technologies = nil
	result.Tags = nil

	// Create the main result record first
	if err := dw.conn.Create(result).Error; err != nil {
		return err
	}

	// Batch size to avoid "too many SQL variables" error
	// SQLite has a default limit of 999 variables
	batchSize := 100

	// Insert associations in batches
	if len(cookies) > 0 {
		// Set ResultID for all cookies
		for i := range cookies {
			cookies[i].ResultID = result.ID
		}
		if err := dw.conn.CreateInBatches(cookies, batchSize).Error; err != nil {
			return err
		}
	}

	if len(headers) > 0 {
		for i := range headers {
			headers[i].ResultID = result.ID
		}
		if err := dw.conn.CreateInBatches(headers, batchSize).Error; err != nil {
			return err
		}
	}

	if len(networkLogs) > 0 {
		for i := range networkLogs {
			networkLogs[i].ResultID = result.ID
		}
		if err := dw.conn.CreateInBatches(networkLogs, batchSize).Error; err != nil {
			return err
		}
	}

	if len(consoleLogs) > 0 {
		for i := range consoleLogs {
			consoleLogs[i].ResultID = result.ID
		}
		if err := dw.conn.CreateInBatches(consoleLogs, batchSize).Error; err != nil {
			return err
		}
	}

	if len(technologies) > 0 {
		for i := range technologies {
			technologies[i].ResultID = result.ID
		}
		if err := dw.conn.CreateInBatches(technologies, batchSize).Error; err != nil {
			return err
		}
	}

	// Handle TLS (single record, not a slice)
	if tls.ID != 0 || tls.Protocol != "" {
		tls.ResultID = result.ID
		if err := dw.conn.Create(&tls).Error; err != nil {
			return err
		}
	}

	// Handle Tags (many-to-many relationship)
	if len(tags) > 0 {
		if err := dw.conn.Model(result).Association("Tags").Append(tags); err != nil {
			return err
		}
	}

	return nil
}

// AssignGroupID assigns a PerceptionHashGroupId based on Hamming distance
func (dw *DbWriter) AssignGroupID(perceptionHashStr string) (uint, error) {
	// Parse the incoming perception hash
	parsedHash, err := islazy.ParsePerceptionHash(perceptionHashStr)
	if err != nil {
		return 0, err
	}

	// Iterate through existing groups to find a match
	for _, group := range dw.hammingGroups {
		dist, err := islazy.HammingDistance(parsedHash, group.Hash)
		if err != nil {
			return 0, err
		}

		if dist <= hammingThreshold {
			return group.GroupID, nil
		}
	}

	// No matching group found; create a new group
	var maxGroupID uint
	err = dw.conn.Model(&models.Result{}).
		Select("COALESCE(MAX(perception_hash_group_id), 0)").
		Scan(&maxGroupID).Error
	if err != nil {
		return 0, err
	}
	nextGroupID := maxGroupID + 1

	// Add the new group to in-memory cache
	newGroup := islazy.HammingGroup{
		GroupID: nextGroupID,
		Hash:    parsedHash,
	}
	dw.hammingGroups = append(dw.hammingGroups, newGroup)

	return nextGroupID, nil
}

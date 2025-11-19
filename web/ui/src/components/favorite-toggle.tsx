import { useState } from "react";
import { Star } from "lucide-react";
import { Button } from "@/components/ui/button";
import * as api from "@/lib/api/api";
import * as apitypes from "@/lib/api/types";
import { cn } from "@/lib/utils";

interface FavoriteToggleProps {
  resultId: number;
  tags: apitypes.tag[];
  onToggle?: () => void;
  className?: string;
}

export function FavoriteToggle({ resultId, tags, onToggle, className }: FavoriteToggleProps) {
  const [isFavorited, setIsFavorited] = useState(() =>
    tags?.some(tag => tag.name === "Favorite") || false
  );
  const [isLoading, setIsLoading] = useState(false);

  const handleToggle = async (e: React.MouseEvent) => {
    e.preventDefault(); // Prevent navigation if this is inside a Link
    e.stopPropagation(); // Prevent event bubbling

    setIsLoading(true);

    try {
      if (isFavorited) {
        await api.post("tagremove", {
          result_id: resultId,
          tag_name: "Favorite"
        });
        setIsFavorited(false);
      } else {
        await api.post("tagadd", {
          result_id: resultId,
          tag_name: "Favorite"
        });
        setIsFavorited(true);
      }

      // Call the optional onToggle callback to refresh data if needed
      onToggle?.();
    } catch (error) {
      console.error("Failed to toggle favorite:", error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Button
      variant="ghost"
      size="icon"
      className={cn("h-8 w-8", className)}
      onClick={handleToggle}
      disabled={isLoading}
      title={isFavorited ? "Remove from favorites" : "Add to favorites"}
    >
      <Star
        className={cn(
          "h-4 w-4 transition-colors",
          isFavorited ? "fill-yellow-400 text-yellow-400" : "text-muted-foreground hover:text-yellow-400"
        )}
      />
    </Button>
  );
}

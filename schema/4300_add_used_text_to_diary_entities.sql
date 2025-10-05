-- Add used_text column to diary_entities table
-- This column stores the actual text used in the diary (entity name or alias)
ALTER TABLE diary_entities ADD COLUMN IF NOT EXISTS used_text TEXT NOT NULL DEFAULT '';

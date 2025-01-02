package models

import (
	"log"

	"gorm.io/gorm"
)

// AddSyncStatusMigration adds the sync_status column to the data table.
func AddSyncStatusMigration(db *gorm.DB) {
	log.Println("Running AddSyncStatusMigration...")

	// Add the ENUM type and the column
	queryAddColumn := `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_type WHERE typname = 'sync_status_enum'
			) THEN
				CREATE TYPE sync_status_enum AS ENUM ('pending', 'success');
			END IF;

			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'data' AND column_name = 'sync_status'
			) THEN
				ALTER TABLE data ADD COLUMN sync_status sync_status_enum DEFAULT 'pending';
			END IF;
		END $$;
	`

	// Execute the query to add the column
	if err := db.Exec(queryAddColumn).Error; err != nil {
		log.Panicf("Failed to add sync_status column: %v", err)
	}

	// Update existing rows to have the default value
	queryUpdateDefault := `
		UPDATE data SET sync_status = 'pending' WHERE sync_status IS NULL;
	`

	if err := db.Exec(queryUpdateDefault).Error; err != nil {
		log.Panicf("Failed to set default values for sync_status: %v", err)
	}

	log.Println("AddSyncStatusMigration completed successfully.")
}

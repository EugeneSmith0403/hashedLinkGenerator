ALTER TABLE 
    auth_sessions  
ADD COLUMN 
    is_verify BOOLEAN NOT NULL DEFAULT FALSE;
CREATE TABLE "redirection" (
    "key" TEXT PRIMARY KEY,
    "url" TEXT NOT NULL,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
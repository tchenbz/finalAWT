-- Books Table
CREATE TABLE IF NOT EXISTS books (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    authors TEXT[] NOT NULL, -- Array of authors
    isbn VARCHAR(13) UNIQUE NOT NULL,
    publication_date DATE NOT NULL,
    genre VARCHAR(100),
    description TEXT,
    average_rating FLOAT DEFAULT 0,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);

-- Reviews Table
CREATE TABLE IF NOT EXISTS reviews (
    id SERIAL PRIMARY KEY,
    book_id INT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    rating INT CHECK (rating >= 1 AND rating <= 5), -- Ratings between 1 and 5
    helpful_count INT DEFAULT 0,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);

-- Reading Lists Table
CREATE TABLE IF NOT EXISTS reading_lists (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) CHECK (status IN ('currently reading', 'completed')) NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);

-- Reading List Books (Association Table for Many-to-Many Relationship)
CREATE TABLE IF NOT EXISTS reading_list_books (
    reading_list_id INT NOT NULL REFERENCES reading_lists(id) ON DELETE CASCADE,
    book_id INT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    PRIMARY KEY (reading_list_id, book_id)
);
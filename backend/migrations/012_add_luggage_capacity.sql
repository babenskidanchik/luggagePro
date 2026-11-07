ALTER TABLE hotels
ADD COLUMN luggage_capacity INTEGER NOT NULL DEFAULT 0;

ALTER TABLE hotels
ADD CONSTRAINT hotels_luggage_capacity_check
CHECK (
    luggage_capacity >= 0);
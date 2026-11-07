ALTER TABLE hotel_applications
ADD COLUMN venue_type VARCHAR(32) NOT NULL DEFAULT 'hotel';

ALTER TABLE hotels
ADD COLUMN venue_type VARCHAR(32) NOT NULL DEFAULT 'hotel';

ALTER TABLE hotel_applications
ADD CONSTRAINT hotel_applications_venue_type_check
CHECK (
    venue_type IN (
        'hotel',
        'hostel',
        'guesthouse',
        'apartment',
        'luggage_shop',
        'other'
    )
);

ALTER TABLE hotels
ADD CONSTRAINT hotels_venue_type_check
CHECK (
    venue_type IN (
        'hotel',
        'hostel',
        'guesthouse',
        'apartment',
        'luggage_shop',
        'other'
    )
);

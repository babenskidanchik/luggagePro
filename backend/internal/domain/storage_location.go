package domain

type VenueType string

const (
	VenueTypeHotel       VenueType = "hotel"
	VenueTypeHostel      VenueType = "hostel"
	VenueTypeGuestHouse  VenueType = "guest_house"
	VenueTypeApartment   VenueType = "apartment"
	VenueTypeLuggageShop VenueType = "luggage_shop"
	VenueTypeOther       VenueType = "other"
)

func (v VenueType) IsValid() bool {
	switch v {
	case VenueTypeHotel,
		VenueTypeHostel,
		VenueTypeGuestHouse,
		VenueTypeApartment,
		VenueTypeLuggageShop,
		VenueTypeOther:
		return true
	default:
		return false
	}
}

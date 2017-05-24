package rexon

const (
	// KeyStartTag identifies in RexSet.RexMap the regexp used to start matching for a new document
	KeyStartTag = "_start_tag"
	// KeyDropTag identifies in RexSet.RexMap the regexp used to stop parsing immediately
	KeyDropTag = "_drop_tag"
	// KeySkipTag identifies in RexSet.RexMap the regexp used to begin skipping data until KeyContinueTag
	KeySkipTag = "_skip_tag"
	// KeyContinueTag identifies in RexSet.RexMap the regexp used to continue matching after KeySkipTag
	KeyContinueTag = "_continue_tag"
	// KeyTypeAll identifies in RexLine/Set.FieldTypes the catch all type for type parsing (TypeInt, TypeFloat, TypeBool, TypeString)
	KeyTypeAll = "_all"
	// TypeInt identifies in RexLine/Set.FieldTypes that the given field should be parsed to int
	TypeInt ValueType = "int"
	// TypeFloat identifies in RexLine/Set.FieldTypes that the given field should be parsed to float
	TypeFloat ValueType = "float"
	// TypeBool identifies in RexLine/Set.FieldTypes that the given field should be parsed to bool
	TypeBool ValueType = "bool"
	// TypeString identifies in RexLine/Set.FieldTypes that the given field should be parsed to string
	TypeString ValueType = "string"
	// Bytes bytes
	Bytes = Unit(1)
	// KBytes KiloBytes
	KBytes = Bytes * 1000
	// MBytes MegaBytes
	MBytes = KBytes * 1000
	// GBytes GigaBytes
	GBytes = MBytes * 1000
	// TBytes TeraBytes
	TBytes = GBytes * 1000
	// PBytes PetaBytes
	PBytes = TBytes * 1000
	// EBytes ExaBytes
	EBytes = PBytes * 1000
	// KiBytes KiloBytes
	KiBytes = Bytes * 1024
	// MiBytes MegaBytes
	MiBytes = KiBytes * 1024
	// GiBytes GigaBytes
	GiBytes = MiBytes * 1024
	// TiBytes TeraBytes
	TiBytes = GBytes * 1024
	// PiBytes PetaBytes
	PiBytes = TBytes * 1024
	// EiBytes ExaBytes
	EiBytes = PBytes * 1024
)

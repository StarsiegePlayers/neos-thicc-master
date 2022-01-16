package file

// adapted from https://stackoverflow.com/a/42718395

const (
	bitExecute = iota << 1
	bitWrite
	bitRead
	shiftUser  = 6
	shiftGroup = 3
	shiftOther = 0
)

const (
	UserRead             = bitRead << shiftUser
	UserWrite            = bitWrite << shiftUser
	UserExecute          = bitExecute << shiftUser
	UserReadWrite        = UserRead | UserWrite
	UserReadWriteExecute = UserReadWrite | UserExecute

	GroupRead             = bitRead << shiftGroup
	GroupWrite            = bitWrite << shiftGroup
	GroupExecute          = bitExecute << shiftGroup
	GroupReadWrite        = GroupRead | GroupWrite
	GroupReadWriteExecute = GroupReadWrite | GroupExecute

	OtherRead             = bitRead << shiftOther
	OtherWrite            = bitWrite << shiftOther
	OtherExecute          = bitExecute << shiftOther
	OtherReadWrite        = OtherRead | OtherWrite
	OtherReadWriteExecute = OtherReadWrite | OtherExecute

	AllRead             = UserRead | GroupRead | OtherRead
	AllWrite            = UserWrite | GroupWrite | OtherWrite
	AllExecute          = UserExecute | GroupExecute | OtherExecute
	AllReadWrite        = AllRead | AllWrite
	AllReadWriteExecute = AllReadWrite | AllExecute
)

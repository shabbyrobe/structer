package invalid

type InvalidType struct {
	DoesNotExist
}

type InvalidField struct {
	Field DoesNotExist
}

package driveapi

type DriveApi interface {
	DoesSheetExist(name string) (bool, error)
}

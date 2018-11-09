package data

// ErrorStruct struct
type ErrorStruct struct {
	Status int
	Detail string
}

// ErrorList variable
var ErrorList = make(map[int]ErrorStruct)

func init() {
	ErrorList[1001] = ErrorStruct{500, "failed query string parse"}
	ErrorList[1003] = ErrorStruct{500, "invalid token"}
	ErrorList[1004] = ErrorStruct{400, "this app can only be used in open channels"}
	ErrorList[1005] = ErrorStruct{405, "unknown command"}
	ErrorList[1100] = ErrorStruct{400, "response url is empty"}
	ErrorList[1110] = ErrorStruct{400, "repository already exists"}
	ErrorList[1111] = ErrorStruct{400, "invalid token"}
	ErrorList[1120] = ErrorStruct{400, "repository not found"}
	ErrorList[1121] = ErrorStruct{400, "access denied"}
	ErrorList[1131] = ErrorStruct{400, "in current release doesn't included tasks"}
	ErrorList[1140] = ErrorStruct{400, "job list is empty"}
	ErrorList[1141] = ErrorStruct{400, "job already running"}
	ErrorList[1142] = ErrorStruct{400, "job should be running"}
}

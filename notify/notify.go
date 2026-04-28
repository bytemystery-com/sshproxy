package notify

type NotifyData struct {
	Index int
	Msg   string
	Err   error
}

func SendNotify(notifyChannel chan<- NotifyData, index int, msg string, err error) bool {
	flag := false
	if notifyChannel != nil {
		data := NotifyData{
			Msg:   msg,
			Err:   err,
			Index: index,
		}
		select {
		case notifyChannel <- data:
			flag = true
		default:
		}
	}
	return flag
}

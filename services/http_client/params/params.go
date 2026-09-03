package params

type PidInformation struct {
	Pid int `json:"pid"`
}

func NewPidInformation(pid int) *PidInformation {
	return &PidInformation{
		Pid: pid,
	}
}

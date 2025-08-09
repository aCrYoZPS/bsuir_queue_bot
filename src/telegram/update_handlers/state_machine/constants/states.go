package constants

const (
	IDLE_STATE = ""
)

const (
	ADMIN_SUBMIT_START_STATE     = "submit"
	ADMIN_SUBMITTING_NAME_STATE  = "submitting_name"
	ADMIN_SUBMITTING_GROUP_STATE = "submitting_group"
	ADMIN_SUBMITTING_PROOF_STATE = "submitting_proof"
	ADMIN_WAITING_STATE          = "admin_waiting"
)

const (
	GROUP_SUBMIT_START_STATE     = "group_submit"
	GROUP_SUBMIT_GROUPNAME_STATE = "groupname"
	GROUP_SUBMIT_NAME_STATE      = "group_name"
	GROUP_WAITING_STATE          = "group_waiting"
)

const (
	LABWORK_SUBMIT_START_STATE   = "lab_submit"
	LABWORK_SUBMIT_WAITING_STATE = "lab_wait"
	LABWORK_SUBMIT_PROOF_STATE   = "lab_proof_submit"
)

const (
	LABWORK_ADD_START_STATE       = "custom_lab_start"
	LABWORK_ADD_SUBMIT_NAME_STATE = "custom_lab_name"
	LABWORK_ADD_WAITING_STATE     = "custom_lab_wait"
)

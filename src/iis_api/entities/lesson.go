package iis_api_entities

// I mean, it's basically a enum.  
type Subgroup int8

const (
	AllSubgroups   = Subgroup(0)
	FirstSubgroup  = Subgroup(1)
	SecondSubgroup = Subgroup(2)
)

type Lesson struct {
	Subject        string   `json:"subject"`
	LessonType     string   `json:"lessonTypeAbbrev"`
	SubgroupNumber Subgroup `json:"numSubgroup"`
	WeekNumber     []int8   `json:"weekNumber"`
}

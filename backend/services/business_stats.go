package services

import (
	"sort"
	"strings"

	"mcloud/database"
	"mcloud/models"
)

const unspecifiedKey = "(未填写)"

type BusinessHost struct {
	Name    string `json:"name"`
	IP      string `json:"ip"`
	Project string `json:"project"`
}

type BusinessGroup struct {
	Key        string         `json:"key"`
	HostCount  int            `json:"host_count"`
	CPU        int            `json:"cpu"`
	Memory     int            `json:"memory"`
	Storage    int            `json:"storage"`
	Running    int            `json:"running"`
	Stopped    int            `json:"stopped"`
	Projects   []string       `json:"projects"`
	Companies  []string       `json:"companies"`
	Applicants []string       `json:"applicants"`
	Hosts      []BusinessHost `json:"hosts"`
}

type BusinessSummary struct {
	HostCount      int `json:"host_count"`
	ProjectCount   int `json:"project_count"`
	CompanyCount   int `json:"company_count"`
	ApplicantCount int `json:"applicant_count"`
}

type BusinessStatsResponse struct {
	Summary     BusinessSummary `json:"summary"`
	ByProject   []BusinessGroup `json:"by_project"`
	ByCompany   []BusinessGroup `json:"by_company"`
	ByApplicant []BusinessGroup `json:"by_applicant"`
}

type businessAgg struct {
	group        BusinessGroup
	projectSet   map[string]bool
	companySet   map[string]bool
	applicantSet map[string]bool
}

func newBusinessAgg(key string) *businessAgg {
	return &businessAgg{
		group:        BusinessGroup{Key: key},
		projectSet:   map[string]bool{},
		companySet:   map[string]bool{},
		applicantSet: map[string]bool{},
	}
}

func (a *businessAgg) add(h models.Host, project, company, applicant string) {
	a.group.HostCount++
	a.group.CPU += h.CPU
	a.group.Memory += h.Memory
	a.group.Storage += h.SystemDisk + h.DataDisk
	if h.Status == "运行中" {
		a.group.Running++
	} else if h.Status == "已停止" || h.Status == "已关机" {
		a.group.Stopped++
	}
	a.group.Hosts = append(a.group.Hosts, BusinessHost{
		Name:    h.Name,
		IP:      h.PrivateIP,
		Project: project,
	})
	if project != unspecifiedKey {
		a.projectSet[project] = true
	}
	if company != unspecifiedKey {
		a.companySet[company] = true
	}
	if applicant != unspecifiedKey {
		a.applicantSet[applicant] = true
	}
}

func (a *businessAgg) finalize() BusinessGroup {
	g := a.group
	g.Projects = sortedKeys(a.projectSet)
	g.Companies = sortedKeys(a.companySet)
	g.Applicants = sortedKeys(a.applicantSet)
	return g
}

func sortedKeys(m map[string]bool) []string {
	out := []string{}
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func normKey(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return unspecifiedKey
	}
	return s
}

func (s *StatsService) GetBusinessStats() (*BusinessStatsResponse, error) {
	var hosts []models.Host
	if err := database.DB.Preload("Application").Find(&hosts).Error; err != nil {
		return nil, err
	}

	projectMap := map[string]*businessAgg{}
	companyMap := map[string]*businessAgg{}
	applicantMap := map[string]*businessAgg{}

	projectSet := map[string]bool{}
	companySet := map[string]bool{}
	applicantSet := map[string]bool{}

	for _, h := range hosts {
		project, company, applicant := unspecifiedKey, unspecifiedKey, unspecifiedKey
		if h.Application != nil {
			project = normKey(h.Application.Project)
			company = normKey(h.Application.ApplyUnit)
			applicant = normKey(h.Application.Applicant)
		}
		if project != unspecifiedKey {
			projectSet[project] = true
		}
		if company != unspecifiedKey {
			companySet[company] = true
		}
		if applicant != unspecifiedKey {
			applicantSet[applicant] = true
		}

		if _, ok := projectMap[project]; !ok {
			projectMap[project] = newBusinessAgg(project)
		}
		projectMap[project].add(h, project, company, applicant)

		if _, ok := companyMap[company]; !ok {
			companyMap[company] = newBusinessAgg(company)
		}
		companyMap[company].add(h, project, company, applicant)

		if _, ok := applicantMap[applicant]; !ok {
			applicantMap[applicant] = newBusinessAgg(applicant)
		}
		applicantMap[applicant].add(h, project, company, applicant)
	}

	return &BusinessStatsResponse{
		Summary: BusinessSummary{
			HostCount:      len(hosts),
			ProjectCount:   len(projectSet),
			CompanyCount:   len(companySet),
			ApplicantCount: len(applicantSet),
		},
		ByProject:   finalizeAggs(projectMap),
		ByCompany:   finalizeAggs(companyMap),
		ByApplicant: finalizeAggs(applicantMap),
	}, nil
}

func finalizeAggs(m map[string]*businessAgg) []BusinessGroup {
	out := make([]BusinessGroup, 0, len(m))
	for _, a := range m {
		out = append(out, a.finalize())
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].HostCount != out[j].HostCount {
			return out[i].HostCount > out[j].HostCount
		}
		if out[i].CPU != out[j].CPU {
			return out[i].CPU > out[j].CPU
		}
		return out[i].Key < out[j].Key
	})
	return out
}

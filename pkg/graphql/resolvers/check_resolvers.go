package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/kfsoftware/statuspage/config"
	"github.com/kfsoftware/statuspage/pkg/check"
	"github.com/kfsoftware/statuspage/pkg/db"
	"github.com/kfsoftware/statuspage/pkg/graphql/generated"
	"github.com/kfsoftware/statuspage/pkg/graphql/models"
	"github.com/kfsoftware/statuspage/pkg/jobs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strings"
	"time"
)

type Resolver struct {
	Db         *gorm.DB
	Registry   *jobs.SchedulerRegistry
	DriverName config.DriverName
}
type httpCheckResolver struct{ *Resolver }

func (h httpCheckResolver) Uptime(ctx context.Context, obj *models.HTTPCheck) (*models.CheckUptime, error) {
	return h.getUptimeForCheck(obj.ID)
}

func (h httpCheckResolver) LatestExecutions(ctx context.Context, obj *models.HTTPCheck, limit int) ([]*models.CheckExecution, error) {
	return h.getLatestExecutions(obj.ID, limit)
}

func (h httpCheckResolver) Executions(ctx context.Context, obj *models.HTTPCheck, from *time.Time, until *time.Time) ([]*models.CheckExecution, error) {
	return h.getExecutions(obj.ID, from, until)
}

func (r *Resolver) HttpCheck() generated.HttpCheckResolver {
	return &httpCheckResolver{r}
}

type icmpCheckResolver struct{ *Resolver }

func (i icmpCheckResolver) Uptime(ctx context.Context, obj *models.ICMPCheck) (*models.CheckUptime, error) {
	return i.getUptimeForCheck(obj.ID)
}

func (i icmpCheckResolver) LatestExecutions(ctx context.Context, obj *models.ICMPCheck, limit int) ([]*models.CheckExecution, error) {
	return i.getLatestExecutions(obj.ID, limit)
}

func (i icmpCheckResolver) Executions(ctx context.Context, obj *models.ICMPCheck, from *time.Time, until *time.Time) ([]*models.CheckExecution, error) {
	return i.getExecutions(obj.ID, from, until)
}

func (r *Resolver) IcmpCheck() generated.IcmpCheckResolver {
	return &icmpCheckResolver{r}
}

type tcpCheckResolver struct{ *Resolver }

func (t tcpCheckResolver) Uptime(ctx context.Context, obj *models.TCPCheck) (*models.CheckUptime, error) {
	return t.getUptimeForCheck(obj.ID)
}

func (t tcpCheckResolver) LatestExecutions(ctx context.Context, obj *models.TCPCheck, limit int) ([]*models.CheckExecution, error) {
	return t.getLatestExecutions(obj.ID, limit)
}

func (t tcpCheckResolver) Executions(ctx context.Context, obj *models.TCPCheck, from *time.Time, until *time.Time) ([]*models.CheckExecution, error) {
	return t.getExecutions(obj.ID, from, until)
}

func (r *Resolver) TcpCheck() generated.TcpCheckResolver {
	return &tcpCheckResolver{r}
}

type tlsCheckResolver struct{ *Resolver }

func (t tlsCheckResolver) Uptime(ctx context.Context, obj *models.TLSCheck) (*models.CheckUptime, error) {
	return t.getUptimeForCheck(obj.ID)
}

type UptimeRawSQL struct {
	Uptimeratio24h float64
	Uptimeratio7d  float64
	Uptimeratio30d float64
}

func (r Resolver) getUptimeForCheck(checkID string) (*models.CheckUptime, error) {
	var uptimeRaw UptimeRawSQL
	switch r.DriverName {
	case config.SQLiteDriver:
		result := r.Db.Raw(`
with t as (select julianday(case
                                when lag(created_at) over (order by created_at desc) is not null
                                    then lag(created_at) over (order by created_at desc)
                                else datetime('now', 'localtime') end) - julianday(created_at) duration,
                  created_at,
                  status,
                  check_id
           from check_execution
           where check_id = ?)
select (
               sum(duration) filter ( where created_at > datetime('now', '-1 days') ) /
               sum(duration) filter ( where status = 'UP' and created_at > datetime('now', '-1 days') )
           ) as uptimeratio24h,
       (
               sum(duration) filter ( where created_at > datetime('now', '-7 days') ) /
               sum(duration) filter ( where status = 'UP' and created_at > datetime('now', '-7 days') )
           ) as uptimeratio7d,
       (
               sum(duration) filter ( where created_at > datetime('now', '-30 days') ) /
               sum(duration) filter ( where status = 'UP' and created_at > datetime('now', '-30 days') )
           ) as uptimeratio30d
from t
group by check_id;
`, checkID).Scan(&uptimeRaw)
		if result.Error != nil {
			return nil, result.Error
		}
	case config.MySQLDriver:
		return nil, errors.New("not implemented")
	case config.PostgresqlDriver:
		result := r.Db.Raw(`
with t as (select case
                      when lag(created_at) over (order by created_at desc) is not null
                          then lag(created_at) over (order by created_at desc)
                      else now() end - created_at duration, created_at, status, check_id
           from check_execution
           where check_id = ?)
select
       sum(duration) filter ( where status = 'UP' and created_at > now() - interval '24 hours')      uptimeduration24h,
       sum(duration) filter ( where created_at > now() - interval '24 hours' )                       totalduration24h,
       extract('epoch' from sum(duration) filter ( where status = 'UP' and created_at > now() - interval '24 hours' )) /
       extract('epoch' from sum(duration) filter ( where created_at > now() - interval '24 hours' )) uptimeratio24h,


       sum(duration) filter ( where status = 'UP' and created_at > now() - interval '7 days')      uptimeduration7d,
       sum(duration) filter ( where created_at > now() - interval '7 days' )                       totalduration7d,
       extract('epoch' from sum(duration) filter ( where status = 'UP' and created_at > now() - interval '7 days' )) /
       extract('epoch' from sum(duration) filter ( where created_at > now() - interval '7 days' )) uptimeratio7d,


       sum(duration) filter ( where status = 'UP' and created_at > now() - interval '30d')      uptimeduration30d,
       sum(duration) filter ( where created_at > now() - interval '30d' )                       totalduration30d,
       extract('epoch' from sum(duration) filter ( where status = 'UP' and created_at > now() - interval '30d' )) /
       extract('epoch' from sum(duration) filter ( where created_at > now() - interval '30d' )) uptimeratio30d,
       check_id
from t
group by check_id;
`, checkID).Scan(&uptimeRaw)
		if result.Error != nil {
			return nil, result.Error
		}
	}
	return &models.CheckUptime{
		Uptime24h: uptimeRaw.Uptimeratio24h,
		Uptime7d:  uptimeRaw.Uptimeratio7d,
		Uptime30d: uptimeRaw.Uptimeratio30d,
	}, nil
}

func (t tlsCheckResolver) LatestExecutions(ctx context.Context, obj *models.TLSCheck, limit int) ([]*models.CheckExecution, error) {
	return t.getLatestExecutions(obj.ID, limit)
}

func (t tlsCheckResolver) Executions(ctx context.Context, obj *models.TLSCheck, from *time.Time, until *time.Time) ([]*models.CheckExecution, error) {
	return t.getExecutions(obj.ID, from, until)
}

func (r *Resolver) TlsCheck() generated.TlsCheckResolver {
	return &tlsCheckResolver{r}
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

func (r *Resolver) StatusPage() generated.StatusPageResolver { return &statusPageResolver{r} }

type statusPageResolver struct{ *Resolver }

func (s statusPageResolver) Checks(ctx context.Context, obj *models.StatusPage) ([]models.Check, error) {
	var pageChecks []db.PageCheck
	result := s.Db.Preload("Check").Preload("Check.Namespace").Order("\"order\" asc").Where("status_page_id = ?", obj.ID).Find(&pageChecks)
	if result.Error != nil {
		return nil, result.Error
	}
	var modelChecks []models.Check
	for _, pageCheck := range pageChecks {
		modelChk, err := mapCheck(pageCheck.Check)
		if err != nil {
			return nil, err
		}
		modelChecks = append(modelChecks, modelChk)
	}
	return modelChecks, nil
}

type mutationResolver struct{ *Resolver }

func (m mutationResolver) DeleteStatusPage(ctx context.Context, name string, namespace string) (*models.DeleteResponse, error) {
	statusPage := db.StatusPage{}
	ns, err := m.createNamespaceIfNotExists(namespace)
	if err != nil {
		return nil, err
	}
	result := m.Db.Find(&statusPage, "name = ? AND namespace_id = ?", name, ns.ID)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "not found") {
			return &models.DeleteResponse{ID: ""}, nil
		}
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return &models.DeleteResponse{ID: ""}, nil
	}
	statusPage.Name = fmt.Sprintf("deleted__%s", statusPage.Name)
	statusPage.Slug = fmt.Sprintf("deleted__%s", statusPage.Slug)
	result = m.Db.Save(&statusPage)
	if result.Error != nil {
		return nil, result.Error
	}
	result = m.Db.Delete(&statusPage)
	if result.Error != nil {
		return nil, result.Error
	}
	return &models.DeleteResponse{ID: statusPage.ID}, nil
}

func (m mutationResolver) DeleteCheck(ctx context.Context, name string, namespace string) (*models.DeleteResponse, error) {
	chk := db.Check{}
	ns, err := m.createNamespaceIfNotExists(namespace)
	if err != nil {
		return nil, err
	}
	result := m.Db.Find(&chk, "name = ? AND namespace_id = ?", name, ns.ID)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "not found") {
			return &models.DeleteResponse{ID: ""}, nil
		}
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return &models.DeleteResponse{ID: ""}, nil
	}
	chk.Name = fmt.Sprintf("deleted__%s", chk.Name)
	result = m.Db.Save(&chk)
	if result.Error != nil {
		return nil, result.Error
	}
	result = m.Db.Delete(&chk)
	if result.Error != nil {
		return nil, result.Error
	}
	var pageChecks []db.PageCheck
	result = m.Db.Where("check_id = ?", chk.ID).Delete(&pageChecks)
	if result.Error != nil {
		return nil, result.Error
	}
	err = m.Registry.Unregister(chk.ID)
	if err != nil {
		return nil, err
	}
	return &models.DeleteResponse{ID: chk.ID}, nil
}

func (m mutationResolver) CreateStatusPage(ctx context.Context, input models.CreateStatusPageInput) (*models.StatusPage, error) {
	data := db.StatusPageData{
		OrderChecks: input.CheckSlugs,
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	ns, err := m.createNamespaceIfNotExists(input.Namespace)
	if err != nil {
		return nil, err
	}
	var checks []db.Check
	result := m.Db.Find(&checks, "name in ? AND namespace_id = ?", input.CheckSlugs, ns.ID)
	if result.Error != nil {
		return nil, result.Error
	}

	statusPage := db.StatusPage{}
	resultDb := m.Db.First(&statusPage, "name = ? AND namespace_id = ?", input.Name, ns.ID)
	if resultDb.Error == nil {
		statusPage.Data = jsonBytes
		statusPage.Title = input.Title
		resultDb = m.Db.Save(statusPage)
		if resultDb.Error != nil {
			return nil, resultDb.Error
		}
		result = m.Db.Where("status_page_id = ?", statusPage.ID).Delete(&db.PageCheck{})
		if result.Error != nil {
			return nil, result.Error
		}
		var pageChecks []db.PageCheck
		for _, chk := range checks {
			pageChecks = append(pageChecks, db.PageCheck{
				CheckID:      chk.ID,
				StatusPageID: statusPage.ID,
				Order: SliceIndex(len(checks), func(i int) bool {
					return chk.Name == input.CheckSlugs[i]
				}),
			})
		}
		result = m.Db.Create(&pageChecks)
		if result.Error != nil {
			return nil, result.Error
		}
		return mapStatusPage(statusPage)
	}
	statusPageSlug := slug.Make(fmt.Sprintf("%s-%s", input.Namespace, input.Name))
	statusPageId := uuid.New().String()
	var pageChecks []db.PageCheck
	for _, chk := range checks {
		pageChecks = append(pageChecks, db.PageCheck{
			CheckID:      chk.ID,
			StatusPageID: statusPageId,
			Order: SliceIndex(len(checks), func(i int) bool {
				return chk.Name == input.CheckSlugs[i]
			}),
		})
	}
	result = m.Db.Create(&db.StatusPage{
		ID:        statusPageId,
		Name:      input.Name,
		Slug:      statusPageSlug,
		Namespace: *ns,
		Data:      jsonBytes,
		Title:     input.Title,
		Checks:    pageChecks,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	return &models.StatusPage{
		ID:        statusPageId,
		Name:      input.Name,
		Namespace: ns.Name,
		Title:     input.Title,
	}, nil
}
func SliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}
func mapStatusPage(page db.StatusPage) (*models.StatusPage, error) {
	data := db.StatusPageData{}
	if err := json.Unmarshal(page.Data, &data); err != nil {
		return nil, err
	}
	return &models.StatusPage{
		ID:             page.ID,
		Name:           page.Name,
		Namespace:      page.Namespace.Name,
		Title:          page.Title,
		StatusPageItem: page,
		Slug:           page.Slug,
	}, nil
}

func (m mutationResolver) Poll(ctx context.Context) (*models.PollResult, error) {
	start := time.Now()
	db.CheckAll(m.Db)
	end := time.Now()
	return &models.PollResult{Took: int(end.Sub(start).Milliseconds())}, nil
}

func (m mutationResolver) CreateTCPCheck(ctx context.Context, input models.CreateTCPCheckInput) (models.Check, error) {
	data := db.TcpCheckData{Address: input.Address}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	_, err = time.ParseDuration(input.Frecuency)
	if err != nil {
		return nil, err
	}
	ns, err := m.createNamespaceIfNotExists(input.Namespace)
	if err != nil {
		return nil, err
	}
	chk := db.Check{}
	resultDb := m.Db.First(&chk, "name = ? AND namespace_id = ?", input.Name, ns.ID)
	if resultDb.Error == nil {
		chk.Data = jsonBytes
		chk.Frecuency = input.Frecuency
		resultDb = m.Db.Save(chk)
		if resultDb.Error != nil {
			return nil, resultDb.Error
		}
		return mapCheck(chk)
	}

	checkId := uuid.New().String()
	result := m.Db.Create(&db.Check{
		ID:        checkId,
		Name:      input.Name,
		Frecuency: input.Frecuency,
		Data:      jsonBytes,
		Type:      check.TcpType,
		Status:    db.Scheduled,
		Namespace: *ns,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	err = m.addCheckResult(checkId)
	if err != nil {
		return nil, err
	}
	return models.TCPCheck{
		ID:        checkId,
		Name:      input.Name,
		Namespace: ns.Name,
		Frecuency: input.Frecuency,
		Address:   input.Address,
	}, nil
}
func (m mutationResolver) addCheckResult(id string) error {
	chk := db.Check{}
	resultDb := m.Db.First(&chk, "id = ?", id)
	if resultDb.Error != nil {
		return resultDb.Error
	}
	interval, err := time.ParseDuration(chk.Frecuency)
	if err != nil {
		return err
	}
	chk.Check(m.Db)
	err = m.Registry.Register(chk.ID, interval, func() {
		chk := db.Check{}
		resultDb := m.Db.First(&chk, "id = ?", id)
		if resultDb.Error != nil {
			log.Warnf("Error getting check %s: %s", id, resultDb.Error)
			return
		}
		chk.Check(m.Db)
	})
	if err != nil {
		return err
	}
	return nil
}
func (r Resolver) createNamespaceIfNotExists(namespace string) (*db.Namespace, error) {
	namespaceId := uuid.New().String()
	ns := &db.Namespace{}
	resultDb := r.Db.First(ns, "name = ?", namespace)
	if resultDb.Error == nil {
		return ns, nil
	}
	result := r.Db.Create(&db.Namespace{
		ID:   namespaceId,
		Name: namespace,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	return &db.Namespace{
		ID:   namespaceId,
		Name: namespace,
	}, nil
}
func (m mutationResolver) CreateTLSCheck(ctx context.Context, input models.CreateTLSCheckInput) (models.Check, error) {
	data := db.TlsCheckData{
		Address: input.Address,
	}
	if input.RootCAs != nil {
		data.RootCAs = *input.RootCAs
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	_, err = time.ParseDuration(input.Frecuency)
	if err != nil {
		return nil, err
	}
	ns, err := m.createNamespaceIfNotExists(input.Namespace)
	if err != nil {
		return nil, err
	}
	chk := db.Check{}
	resultDb := m.Db.First(&chk, "name = ? AND namespace_id = ?", input.Name, ns.ID)
	if resultDb.Error == nil {
		chk.Data = jsonBytes
		chk.Frecuency = input.Frecuency
		resultDb = m.Db.Save(chk)
		if resultDb.Error != nil {
			return nil, resultDb.Error
		}
		return mapCheck(chk)
	}
	checkId := uuid.New().String()
	result := m.Db.Create(&db.Check{
		ID:        checkId,
		Name:      input.Name,
		Frecuency: input.Frecuency,
		Data:      jsonBytes,
		Namespace: *ns,
		Type:      check.TlsType,
		Status:    db.Scheduled,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	err = m.addCheckResult(checkId)
	if err != nil {
		return nil, err
	}
	return models.TLSCheck{
		ID:        checkId,
		Name:      input.Name,
		Namespace: ns.Name,
		Frecuency: input.Frecuency,
		Address:   input.Address,
	}, nil
}

func (m mutationResolver) CreateICMPCheck(ctx context.Context, input models.CreateICMPCheckInput) (models.Check, error) {
	data := db.IcmpCheckData{Address: input.Address}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	_, err = time.ParseDuration(input.Frecuency)
	if err != nil {
		return nil, err
	}
	chk := db.Check{}
	ns, err := m.createNamespaceIfNotExists(input.Namespace)
	if err != nil {
		return nil, err
	}
	resultDb := m.Db.First(&chk, "name = ? AND namespace_id = ?", input.Name, ns.ID)
	if resultDb.Error == nil {
		chk.Data = jsonBytes
		chk.Frecuency = input.Frecuency
		resultDb = m.Db.Save(chk)
		if resultDb.Error != nil {
			return nil, resultDb.Error
		}
		return mapCheck(chk)
	}
	checkId := uuid.New().String()
	result := m.Db.Create(&db.Check{
		ID:        checkId,
		Name:      input.Name,
		Frecuency: input.Frecuency,
		Data:      jsonBytes,
		Type:      check.IcmpType,
		Status:    db.Scheduled,
		Namespace: *ns,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	err = m.addCheckResult(checkId)
	if err != nil {
		return nil, err
	}
	return models.ICMPCheck{
		ID:        checkId,
		Name:      input.Name,
		Namespace: ns.Name,
		Frecuency: input.Frecuency,
		Address:   input.Address,
	}, nil
}

func (m mutationResolver) CreateHTTPCheck(ctx context.Context, input models.CreateHTTPCheckInput) (models.Check, error) {
	data := db.HttpCheckData{Url: input.URL, StatusCode: input.StatusCode}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	_, err = time.ParseDuration(input.Frecuency)
	if err != nil {
		return nil, err
	}
	ns, err := m.createNamespaceIfNotExists(input.Namespace)
	if err != nil {
		return nil, err
	}
	chk := db.Check{}
	resultDb := m.Db.First(&chk, "name = ? AND namespace_id = ?", input.Name, ns.ID)
	if resultDb.Error == nil {
		chk.Data = jsonBytes
		chk.Frecuency = input.Frecuency
		resultDb = m.Db.Save(chk)
		if resultDb.Error != nil {
			return nil, resultDb.Error
		}
		return mapCheck(chk)
	}
	checkId := uuid.New().String()
	result := m.Db.Create(&db.Check{
		ID:        checkId,
		Name:      input.Name,
		Frecuency: input.Frecuency,
		Data:      jsonBytes,
		Namespace: *ns,
		Type:      check.HttpType,
		Status:    db.Scheduled,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	err = m.addCheckResult(checkId)
	if err != nil {
		return nil, err
	}
	return models.HTTPCheck{
		ID:        checkId,
		Name:      input.Name,
		Namespace: ns.Name,
		Frecuency: input.Frecuency,
		URL:       input.URL,
	}, nil
}

type queryResolver struct{ *Resolver }

func (q queryResolver) StatusPages(ctx context.Context, namespace *string) ([]*models.StatusPage, error) {
	var statusPages []db.StatusPage
	if namespace == nil {
		result := q.Db.Preload("Namespace").Find(&statusPages)
		if result.Error != nil {
			return nil, result.Error
		}
	} else {
		ns := db.Namespace{}
		result := q.Db.First(&ns, "name = ?", *namespace)
		if result.Error != nil {
			return nil, errors.Errorf("namespace %s not found", *namespace)
		}
		result = q.Db.Preload("Namespace").Find(&statusPages, "namespace_id = ?", ns.ID)
		if result.Error != nil {
			return nil, result.Error
		}
	}
	var modelStatusPages []*models.StatusPage
	for _, chk := range statusPages {
		modelChk, err := mapStatusPage(chk)
		if err != nil {
			return nil, err
		}
		modelStatusPages = append(modelStatusPages, modelChk)
	}
	return modelStatusPages, nil
}

func (q queryResolver) StatusPage(ctx context.Context, slug string) (*models.StatusPage, error) {
	statusPage := db.StatusPage{}
	result := q.Db.Preload("Namespace").First(&statusPage, "slug = ?", slug)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "not found") {
			return nil, nil
		}
		return nil, result.Error
	}
	return mapStatusPage(statusPage)
}

func (q queryResolver) Namespaces(ctx context.Context) ([]*models.Namespace, error) {
	var dbNamespaces []*db.Namespace
	result := q.Db.Find(&dbNamespaces)
	if result.Error != nil {
		return nil, result.Error
	}
	var nsModels []*models.Namespace
	for _, ns := range dbNamespaces {
		nsModels = append(nsModels, mapNamespace(ns))
	}
	return nsModels, nil
}

func mapNamespace(n *db.Namespace) *models.Namespace {
	return &models.Namespace{
		ID:   n.ID,
		Name: n.Name,
	}
}
func (r Resolver) getLatestExecutions(checkID string, limit int) ([]*models.CheckExecution, error) {
	var executions []db.CheckExecution
	result := r.Db.Order("created_at asc").
		Where("check_id = ?", checkID).Limit(limit).
		Find(&executions)
	if result.Error != nil {
		return nil, result.Error
	}
	var modelExecutions []*models.CheckExecution
	for _, execution := range executions {
		modelExecutions = append(modelExecutions, mapCheckExecution(execution))
	}
	return modelExecutions, nil
}
func (q queryResolver) LatestExecutions(ctx context.Context, checkID string, limit int) ([]*models.CheckExecution, error) {
	return q.getLatestExecutions(checkID, limit)
}

func (q queryResolver) Check(ctx context.Context, checkID string) (models.Check, error) {
	chk := db.Check{}
	result := q.Db.First(&chk, "id = ?", checkID)
	if result.Error != nil {
		return nil, result.Error
	}
	return mapCheck(chk)
}

func (q queryResolver) Execution(ctx context.Context, execID string) (*models.CheckExecution, error) {
	chkExec := db.CheckExecution{}
	result := q.Db.First(&chkExec, "id = ?", execID)
	if result.Error != nil {
		return nil, result.Error
	}
	return mapCheckExecution(chkExec), nil
}

func (r Resolver) getExecutions(checkID string, from *time.Time, until *time.Time) ([]*models.CheckExecution, error) {
	var executions []db.CheckExecution
	result := r.Db.Order("created_at desc").Where("check_id = ? AND created_at BETWEEN ? AND ?", checkID, from, until).Find(&executions)
	if result.Error != nil {
		return nil, result.Error
	}
	var modelExecutions []*models.CheckExecution
	for _, execution := range executions {
		modelExecutions = append(modelExecutions, mapCheckExecution(execution))
	}
	return modelExecutions, nil
}
func (q queryResolver) Executions(ctx context.Context, checkID string, from *time.Time, until *time.Time) ([]*models.CheckExecution, error) {
	var executions []db.CheckExecution
	result := q.Db.Order("created_at desc").Where("check_id = ? AND created_at BETWEEN ? AND ?", checkID, from, until).Find(&executions)
	if result.Error != nil {
		return nil, result.Error
	}
	var modelExecutions []*models.CheckExecution
	for _, execution := range executions {
		modelExecutions = append(modelExecutions, mapCheckExecution(execution))
	}
	return modelExecutions, nil

}

func (q queryResolver) Checks(ctx context.Context, namespace *string) ([]models.Check, error) {
	var checks []db.Check
	if namespace == nil {
		result := q.Db.Preload("Namespace").Find(&checks)
		if result.Error != nil {
			return nil, result.Error
		}
	} else {
		ns := db.Namespace{}
		result := q.Db.First(&ns, "name = ?", *namespace)
		if result.Error != nil {
			return nil, errors.Errorf("namespace %s not found", *namespace)
		}
		result = q.Db.Preload("Namespace").Find(&checks, "namespace_id = ?", ns.ID)
		if result.Error != nil {
			return nil, result.Error
		}
	}
	var modelChecks []models.Check
	for _, chk := range checks {
		modelChk, err := mapCheck(chk)
		if err != nil {
			return nil, err
		}
		modelChecks = append(modelChecks, modelChk)
	}
	return modelChecks, nil
}
func mapCheckExecution(chkExec db.CheckExecution) *models.CheckExecution {
	return &models.CheckExecution{
		ID:            chkExec.ID,
		ExecutionTime: chkExec.CreatedAt,
		Message:       chkExec.Message,
		ErrorMsg:      chkExec.ErrorMsg,
		Status:        string(chkExec.Status),
	}
}
func mapCheck(chk db.Check) (models.Check, error) {
	var modelCheck models.Check
	errorMsg := chk.ErrorMsg
	msg := chk.Message
	latestCheck := chk.LatestCheck

	switch chk.Type {
	case check.HttpType:
		httpCheckData, err := chk.GetHttpData()
		if err != nil {
			return nil, err
		}
		modelCheck = models.HTTPCheck{
			ID:          chk.ID,
			Name:        chk.Name,
			Namespace:   chk.Namespace.Name,
			Frecuency:   chk.Frecuency,
			URL:         httpCheckData.Url,
			Status:      string(chk.Status),
			LatestCheck: &latestCheck,
			ErrorMsg:    errorMsg,
			Message:     msg,
		}
	case check.TcpType:
		tcpCheckData, err := chk.GetTcpData()
		if err != nil {
			return nil, err
		}
		modelCheck = models.TCPCheck{
			ID:          chk.ID,
			Name:        chk.Name,
			Namespace:   chk.Namespace.Name,
			Frecuency:   chk.Frecuency,
			Address:     tcpCheckData.Address,
			Status:      string(chk.Status),
			LatestCheck: &latestCheck,
			ErrorMsg:    errorMsg,
			Message:     msg,
		}
	case check.TlsType:
		tlsCheckData, err := chk.GetTlsData()
		if err != nil {
			return nil, err
		}
		modelCheck = models.TLSCheck{
			ID:          chk.ID,
			Name:        chk.Name,
			Namespace:   chk.Namespace.Name,
			Frecuency:   chk.Frecuency,
			Address:     tlsCheckData.Address,
			Status:      string(chk.Status),
			LatestCheck: &latestCheck,
			ErrorMsg:    errorMsg,
			Message:     msg,
		}
	case check.IcmpType:
		icmpCheckData, err := chk.GetIcmpData()
		if err != nil {
			return nil, err
		}
		modelCheck = models.TCPCheck{
			ID:          chk.ID,
			Name:        chk.Name,
			Namespace:   chk.Namespace.Name,
			Frecuency:   chk.Frecuency,
			Address:     icmpCheckData.Address,
			Status:      string(chk.Status),
			LatestCheck: &latestCheck,
			ErrorMsg:    errorMsg,
			Message:     msg,
		}
	}
	return modelCheck, nil
}

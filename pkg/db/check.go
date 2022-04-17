package db

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/kfsoftware/statuspage/pkg/check"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"net/http"
	"sync"
	"time"
)

type Status string

const (
	Up        Status = "UP"
	Scheduled Status = "SCHEDULED"
	Checking  Status = "CHECKING"
	Down      Status = "DOWN"
)

type Namespace struct {
	ID        string `gorm:"primary_key"`
	Name      string `gorm:"uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
type PageCheck struct {
	CheckID      string     `gorm:"primaryKey"`
	Check        Check      `gorm:"foreignKey:CheckID"`
	StatusPageID string     `gorm:"primaryKey;uniqueIndex:pagecheckorder;index;not null"`
	StatusPage   StatusPage `gorm:"foreignKey:StatusPageID"`
	Order        int        `gorm:"uniqueIndex:pagecheckorder;index"`
	CreatedAt    time.Time
}
type StatusPage struct {
	ID          string `gorm:"primary_key"`
	Title       string
	Name        string    `gorm:"uniqueIndex:statusslugns;index;not null"`
	Slug        string    `gorm:"uniqueIndex"`
	NamespaceID string    `gorm:"uniqueIndex:statusslugns;index;not null"`
	Namespace   Namespace `gorm:"foreignKey:NamespaceID"`
	Data        datatypes.JSON
	Checks      []PageCheck `gorm:"foreignKey:StatusPageID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (StatusPage) TableName() string {
	return "statuspage"
}
func (s StatusPage) GetData() (*StatusPageData, error) {
	marshalJSON, err := s.Data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	statusPageData := &StatusPageData{}
	err = json.Unmarshal(marshalJSON, &statusPageData)
	if err != nil {
		return nil, err
	}
	return statusPageData, nil
}

type Check struct {
	ID                  string    `gorm:"primaryKey"`
	Name                string    `gorm:"uniqueIndex:checkname;index;not null"`
	NamespaceID         string    `gorm:"uniqueIndex:checkname;index;not null"`
	Namespace           Namespace `gorm:"foreignKey:NamespaceID"`
	Type                check.Type
	Data                datatypes.JSON
	Frecuency           string
	Status              Status
	ErrorMsg            string
	Uptime24h           float64
	Uptime7d            float64
	Uptime30d           float64
	Message             string
	LatestCheck         time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt `gorm:"index"`
	LastFailureNotified *time.Time
	FailureCount        int
	StatusPages         []PageCheck `gorm:"foreignKey:CheckID"`
	Executions          []CheckExecution
}

func (c Check) GetIcmpData() (*IcmpCheckData, error) {
	marshalJSON, err := c.Data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	icmpCheckData := &IcmpCheckData{}
	err = json.Unmarshal(marshalJSON, &icmpCheckData)
	if err != nil {
		return nil, err
	}
	return icmpCheckData, nil
}

func (c Check) GetHttpData() (*HttpCheckData, error) {
	marshalJSON, err := c.Data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	httpCheckData := &HttpCheckData{}
	err = json.Unmarshal(marshalJSON, &httpCheckData)
	if err != nil {
		return nil, err
	}
	return httpCheckData, nil
}

func (c Check) GetTcpData() (*TcpCheckData, error) {
	marshalJSON, err := c.Data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	tcpCheckData := &TcpCheckData{}
	err = json.Unmarshal(marshalJSON, &tcpCheckData)
	if err != nil {
		return nil, err
	}
	return tcpCheckData, nil
}

func (c Check) GetTlsData() (*TlsCheckData, error) {
	marshalJSON, err := c.Data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	tlsCheckData := &TlsCheckData{}
	err = json.Unmarshal(marshalJSON, &tlsCheckData)
	if err != nil {
		return nil, err
	}
	return tlsCheckData, nil
}

func (Check) TableName() string {
	return "check"
}

type CheckExecution struct {
	ID        string `gorm:"primaryKey"`
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
	ErrorMsg  string
	Message   string
	Stats     datatypes.JSON
	CheckID   string
	Check     Check `gorm:"foreignKey:CheckID"`
}

func (CheckExecution) TableName() string {
	return "check_execution"
}

type StatusPageData struct {
	OrderChecks []string
}
type HttpCheckData struct {
	Url        string `json:"url"`
	StatusCode int    `json:"status_code"`
}
type TlsCheckData struct {
	Address string `json:"address"`
	RootCAs string `json:"root_cas"`
}
type TcpCheckData struct {
	Address string `json:"address"`
}
type IcmpCheckData struct {
	Address string `json:"address"`
}

func notifyEndpointUp(db *gorm.DB, chk Check) {
	slackWebhook := viper.GetString("slack.webhook")
	if slackWebhook != "" && chk.LastFailureNotified != nil {
		data := map[string]string{}
		data["text"] = fmt.Sprintf("Endpoint up: %s\n%s", chk.Name, chk.ErrorMsg)
		dataBytes, err := json.Marshal(data)
		if err != nil {
			log.Warnf("Error sending notification to slack:%v", err)
			return
		}
		_, err = http.Post(slackWebhook, "application/json", bytes.NewBuffer(dataBytes))
		if err != nil {
			log.Warnf("Error sending notification to slack:%v", err)
			return
		}
	}
	chk.LastFailureNotified = nil
	chk.FailureCount = 0
	result := db.Save(chk)
	if result.Error != nil {
		log.Warnf("Error saving item :%v", result.Error)
		return
	}
}
func notifyEndpointDown(db *gorm.DB, chk Check) {
	slackWebhook := viper.GetString("slack.webhook")
	if chk.LastFailureNotified != nil && !time.Now().After(chk.LastFailureNotified.Add(5*time.Minute)) {
		log.Infof("Skip notification since lastFailureNotification was=%v", chk.LastFailureNotified)
		return
	}
	if slackWebhook != "" {
		data := map[string]string{}
		data["text"] = fmt.Sprintf("Endpoint down: %s\n%s", chk.Name, chk.ErrorMsg)
		dataBytes, err := json.Marshal(data)
		if err != nil {
			log.Warnf("Error sending notification to slack:%v", err)
			return
		}
		_, err = http.Post(slackWebhook, "application/json", bytes.NewBuffer(dataBytes))
		if err != nil {
			log.Warnf("Error sending notification to slack:%v", err)
			return
		}
	}
	now := time.Now()
	chk.LastFailureNotified = &now
	chk.FailureCount += 1
	result := db.Save(chk)
	if result.Error != nil {
		log.Errorf("Failed to save check=%v with error=%v", chk, result.Error)
	}
}

func CheckAll(db *gorm.DB) {
	start := time.Now()
	err := checkAll(db)
	end := time.Now()
	took := end.Sub(start)
	if err != nil {
		log.Warnf("Failed checking the endpoints: %v", err)
	} else {
		log.Infof("Check executed successfully in %s", took)
	}
	checkUrl := viper.GetString("check.url")
	if checkUrl != "" {
		_, err := http.Get(checkUrl)
		if err != nil {
			log.Errorf("Failed invoking url: %s", checkUrl)
		}
	}
}
func (c *Check) Check(db *gorm.DB) {
	err := c.check(db)
	if err != nil {
		log.Warnf("Failed to verify check=%v", err)
	} else {
		log.Infof("Verified check successfully %s", c.Name)
	}
}
func (c *Check) check(db *gorm.DB) error {
	chk := Check{}
	resultDb := db.First(&chk, "id = ?", c.ID)
	if resultDb.Error != nil {
		return resultDb.Error
	}
	chk.Status = Checking
	resultDb = db.Save(chk)
	if resultDb.Error != nil {
		return resultDb.Error
	}
	var healthChk check.Check
	switch chk.Type {
	case check.HttpType:
		httpCheckData, err := chk.GetHttpData()
		if err != nil {
			log.Errorf("Failed to get http data:%v", err)
			return err
		}
		url := httpCheckData.Url
		var expectedStatusCode int
		if httpCheckData.StatusCode == 0 {
			expectedStatusCode = 200
		} else {
			expectedStatusCode = httpCheckData.StatusCode
		}
		healthChk = check.NewHttpCheck(url, &expectedStatusCode)
	case check.TlsType:
		tlsCheckData, err := chk.GetTlsData()
		if err != nil {
			log.Errorf("Failed to get http data:%v", err)
			return err
		}
		address := tlsCheckData.Address
		var tlsConfig *tls.Config
		if tlsCheckData.RootCAs != "" {
			rootCAs, _ := x509.SystemCertPool()
			if rootCAs == nil {
				rootCAs = x509.NewCertPool()
			}
			ok := rootCAs.AppendCertsFromPEM([]byte(tlsCheckData.RootCAs))
			if !ok {
				log.Warnf("Root CAs not valid: %v", ok)
			}
			tlsConfig = &tls.Config{
				VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
					return nil
				},
				VerifyConnection: func(state tls.ConnectionState) error {
					return nil
				},
				RootCAs: rootCAs,
			}
		}
		healthChk = check.NewTlsCheck(address, tlsConfig)
	case check.IcmpType:
		icmpCheckData, err := chk.GetIcmpData()
		if err != nil {
			log.Errorf("Failed to get http data:%v", err)
			return err
		}
		address := icmpCheckData.Address
		healthChk = check.NewIcmpCheck(address)
	case check.TcpType:
		tlsCheckData, err := chk.GetTcpData()
		if err != nil {
			log.Errorf("Failed to get http data:%v", err)
			return err
		}
		address := tlsCheckData.Address
		healthChk = check.NewTcpCheck(address)
	}
	if healthChk != nil {
		result := healthChk.Check()
		var status Status
		if result.Error != nil {
			status = Down
		} else {
			status = Up
		}
		chk.Status = status
		if result.Error != nil {
			chk.ErrorMsg = result.Error.Error()
		} else {
			chk.ErrorMsg = ""
		}
		chk.Message = result.Message
		chk.LatestCheck = time.Now()
		statsBytes, err := json.Marshal(result.Statistics)
		if err != nil {
			log.Errorf("Health check failed id=%s type=%s err=%v", chk.ID, chk.Type, err)
			return err
		}
		chkExecution := CheckExecution{
			ID:     uuid.New().String(),
			Status: status,
			Stats:  statsBytes,
			Check:  chk,
		}
		resultDb := db.Create(&chkExecution)
		if resultDb.Error != nil {
			log.Errorf("Health check failed id=%s type=%s err=%v", chk.ID, chk.Type, resultDb.Error)
			return result.Error
		}

		resultDb = db.Save(&chk)
		if resultDb.Error != nil {
			log.Errorf("Failed to save check id=%s type=%s err=%v", chk.ID, chk.Type, resultDb.Error)
			return result.Error
		}
		if chk.Status == Down {
			notifyEndpointDown(db, chk)
		} else {
			notifyEndpointUp(db, chk)
		}
	} else {
		log.Warnf("No healthcheck found for id=%s type=%s", chk.ID, string(chk.Type))
	}
	checkUrl := viper.GetString("check.url")
	if checkUrl != "" {
		_, err := http.Get(checkUrl)
		if err != nil {
			log.Errorf("Failed invoking url: %s", checkUrl)
		}
	}
	return nil
}

func checkAll(db *gorm.DB) error {
	var checks []Check
	result := db.Find(&checks)
	if result.Error != nil {
		return result.Error
	}
	var wg sync.WaitGroup
	wg.Add(len(checks))
	for _, chk := range checks {
		chk1 := chk
		go func() {
			defer wg.Done()
			chk1.Check(db)
		}()
	}
	wg.Wait()
	return nil
}

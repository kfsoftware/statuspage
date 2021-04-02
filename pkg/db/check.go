package db

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	uuid "github.com/google/uuid"
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

type Check struct {
	ID          string `gorm:"primaryKey"`
	Identifier  string `gorm:"uniqueIndex"`
	Type        check.Type
	Data        datatypes.JSON
	Frecuency   string
	Status      Status
	ErrorMsg    string
	Message     string
	LatestCheck time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Executions  []CheckExecution
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
}

func (CheckExecution) TableName() string {
	return "check_execution"
}

type HttpCheckData struct {
	Url string `json:"url"`
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

func notifyEndpointDown(chk Check) {
	slackWebhook := viper.GetString("slack.webhook")
	if slackWebhook != "" {
		data := map[string]string{}
		data["text"] = fmt.Sprintf("Endpoint down: %s\n%s", chk.Identifier, chk.ErrorMsg)
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
func checkAll(db *gorm.DB) error {
	var checks []Check
	result := db.Find(&checks)
	if result.Error != nil {
		return result.Error
	}
	var wg sync.WaitGroup
	wg.Add(len(checks))
	for _, chk := range checks {
		chk.Status = Checking
		db.Save(chk)
		var healthChk check.Check
		switch chk.Type {
		case check.HttpType:
			httpCheckData, err := chk.GetHttpData()
			if err != nil {
				log.Errorf("Failed to get http data:%v", err)
				continue
			}
			url := httpCheckData.Url
			expectedStatusCode := 200
			healthChk = check.NewHttpCheck(url, &expectedStatusCode)
		case check.TlsType:
			tlsCheckData, err := chk.GetTlsData()
			if err != nil {
				log.Errorf("Failed to get http data:%v", err)
				continue
			}
			address := tlsCheckData.Address
			rootCAs, _ := x509.SystemCertPool()
			if rootCAs == nil {
				rootCAs = x509.NewCertPool()
			}
			if tlsCheckData.RootCAs != "" {
				ok := rootCAs.AppendCertsFromPEM([]byte(tlsCheckData.RootCAs))
				if !ok {
					log.Warnf("Root CAs not valid: %v", ok)
				}
			}
			tlsConfig := &tls.Config{
				VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
					return nil
				},
				VerifyConnection: func(state tls.ConnectionState) error {
					return nil
				},
				RootCAs: rootCAs,
			}
			healthChk = check.NewTlsCheck(address, tlsConfig)
		case check.IcmpType:
			icmpCheckData, err := chk.GetIcmpData()
			if err != nil {
				log.Errorf("Failed to get http data:%v", err)
				continue
			}
			address := icmpCheckData.Address
			healthChk = check.NewIcmpCheck(address)
		case check.TcpType:
			tlsCheckData, err := chk.GetTcpData()
			if err != nil {
				log.Errorf("Failed to get http data:%v", err)
				continue
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
			statsBytes, err := json.Marshal(result.Statistics)
			if err != nil {
				log.Errorf("Health check failed id=%s type=%s err=%v", chk.ID, chk.Type, err)
				continue
			}
			chkExecution := CheckExecution{
				ID:      uuid.New().String(),
				Status:  status,
				Stats:   statsBytes,
				CheckID: chk.ID,
			}
			resultDb := db.Create(&chkExecution)
			if resultDb.Error != nil {
				log.Errorf("Health check failed id=%s type=%s err=%v", chk.ID, chk.Type, resultDb.Error)
			}
			chk.Status = status
			if result.Error != nil {
				chk.ErrorMsg = result.Error.Error()
			}
			chk.Message = result.Message
			chk.LatestCheck = time.Now()
			resultDb = db.Save(&chk)
			if resultDb.Error != nil {
				log.Errorf("Failed to save check id=%s type=%s err=%v", chk.ID, chk.Type, resultDb.Error)
			}
			if chk.Status == Down {
				chkToNotify := chk
				go func() {
					notifyEndpointDown(chkToNotify)
				}()
			}
		} else {
			log.Warnf("No healthcheck found for id=%s type=%s", chk.ID, string(chk.Type))
		}
	}
	return nil
}

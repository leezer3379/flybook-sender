package esc

import (
	"context"
	"github.com/leezer3379/flybook-sender/dataobj"
	"strings"
	"time"
	"fmt"
)

var ET = map[string]string{
	"alert":    "告警",
	"recovery": "恢复",
}

func parseEtime(etime int64) string {
	t := time.Unix(etime, 0)
	return t.Format("2006-01-02 15:04:05")
}

func parseUsers(recvUsers []*dataobj.RecvUser) string {
	var users string
	users = ""
	for _,user := range recvUsers {
		users += user.Username
	}
	return users
}
func PutData(message *dataobj.Message) error {
	var data = make(map[string]interface{})
	data["Status"] = ET[message.Event.EventType]
	data["Sname"] = message.Event.Sname
	data["Endpoint"] = message.ReadableEndpoint
	data["Metric"] = strings.Join(message.Metrics, ",")
	data["Tags"] = message.ReadableTags
	data["Value"] = message.Event.Value
	data["Info"] = message.Event.Info
	data["Etime"] = parseEtime(message.Event.Etime)
	data["Elink"] = message.EventLink
	data["Priority"] = message.Event.Priority
	data["Users"] = parseUsers(message.Event.RecvUser)
	data["@timestamp"] = time.Now()
	put1, err := client.Index().
		Index(cfg.Es.Index).
		Type(cfg.Es.Index).
		BodyJson(data).
		Do(context.Background())
	if err != nil {
		// Handle error
		return err
	}
	fmt.Printf("Indexed tweet %s to index %s, type %s\n", put1.Id, put1.Index, put1.Type)
	return nil
}


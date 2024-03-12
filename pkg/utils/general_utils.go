package utils

import (
	"io"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/libsignal/logger"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"golang.org/x/net/html"
)

var DbUrl string

func container() (*sqlstore.Container, error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New("sqlite3", DbUrl, dbLog)
	return container, err
}

func NewDevice() (*store.Device, error) {
	container, err := container()
	if err != nil {
		return nil, err
	}

	clientName := "CodeChat"

	store.DeviceProps.Os = &clientName
	device := container.NewDevice()
	
	return device, nil
}

func LoadDeviceWa(phoneNumber *string) (*store.Device, error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New("sqlite3", DbUrl, dbLog)

	if err != nil {
		return nil, err
	}

	clientName := "CodeChat"

	store.DeviceProps.Os = &clientName
	var deviceStore *store.Device

	devices, err := container.GetAllDevices()
	if err != nil {
		logger.Warning("Device not found: ", err)
	}

	for _, v := range devices {
		jid := v.ID.String()
		if strings.Contains(jid, *phoneNumber) {
			deviceStore = v
			break
		}
	}

	return deviceStore, nil
}

func LoadAllDevicesWa() ([]*store.Device, error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New("sqlite3", DbUrl, dbLog)

	if err != nil {
		return nil, err
	}

	devices, err := container.GetAllDevices()
	if err != nil {
		logger.Warning("Device not found: ", err)
	}

	return devices, nil
}

func StringJoin(sep string, values ...string) string {
	var build strings.Builder
	for i, v := range values {
		build.WriteString(v)
		if sep != "" && i != len(values)-1 {
			build.WriteString(sep)
		}
	}
	return build.String()
}

// image
// image:alt
// image:width
// image:height
// site_name
// type
// title
// url
// description
func GeneratePreview(link string) (map[string]string, error) {
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	var exec func(*html.Node)

	preview := make(map[string]string)

	exec = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			property := ""
			content := ""
			for _, a := range n.Attr {
				if a.Key == "property" && strings.HasPrefix(a.Val, "og:") {
					property = strings.Replace(a.Val, "og:", "", 1)
				}
				if a.Key == "content" {
					content = a.Val
				}
			}
			if property != "" && content != "" {
				preview[property] = content
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			exec(c)
		}
	}

	exec(doc)

	return preview, nil
}

func GetMediaByte(url string) ([]byte, http.Header, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()

	headers := response.Header

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, err
	}

	return bytes, headers, nil
}

func GetHeadUrl(url string) (http.Header, error) {
	response, err := http.Head(url)
	if err != nil {
		return nil, err
	}

	return response.Header, nil
}

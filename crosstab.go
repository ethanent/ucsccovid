package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type DocID string

const (
	DocIDTesting            DocID = ""
	DocIDDailyAndTotalCases DocID = "97CFAE3E-575C-473D-BFB2-99D65EE39AE5"
)

type CrosstabExportResp struct {
	VQLCmdResponse struct {
		LayoutStatus struct {
			ApplicationPresModel struct {
				PresentationLayerNotification []*struct {
					PresModelHolder struct {
						GenExportFilePresModel struct {
							ResultKey string `json:"resultKey"`
						} `json:"genExportFilePresModel"`
					} `json:"presModelHolder"`
				} `json:"presentationLayerNotification"`
			} `json:"applicationPresModel"`
		} `json:"layoutStatus"`
	} `json:"vqlCmdResponse"`
}

func GetSessionID(c *http.Client) (string, error) {
	resp, err := c.Get("https://visualizedata.ucop.edu/t/UCSCpublic/views/COVID-19DashboardV2/COVID-19Dashboard?:embed=y&:showVizHome=no&:host_url=https://visualizedata.ucop.edu/&:embed_code_version=3&:tabs=no&:toolbar=yes&:showAppBanner=false&:display_spinner=no&:loadOrderID=0")

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	id := resp.Header.Get("X-Session-Id")

	if id == "" {
		return "", errors.New("missing session ID")
	}

	return id, nil
}

func CreateCrosstabCSVRequest(c *http.Client, sessionID string, docID DocID) (requestID string, err error) {
	bodyStr := `--Eyw1WPdq
Content-Disposition: form-data; name="sheetdocId"

{` + string(docID) + `}
--Eyw1WPdq
Content-Disposition: form-data; name="useTabs"

true
--Eyw1WPdq
Content-Disposition: form-data; name="sendNotifications"

true
--Eyw1WPdq--
`

	u, err := url.Parse("https://visualizedata.ucop.edu/vizql/t/UCSCpublic/w/COVID-19DashboardV2/v/COVID-19Dashboard/sessions/")

	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, sessionID + "/commands/tabsrv/export-crosstab-to-csvserver")

	h := http.Header{}

	h.Set("Content-Type", "multipart/form-data; boundary=Eyw1WPdq")
	h.Set("X-Tsi-Active-Tab", "COVID-19%20Dashboard")
	h.Set("X-Tsi-Supports-Accepted", "true")
	h.Set("Content-Length", strconv.Itoa(len([]byte(bodyStr))))

	req := &http.Request{
		Method: "POST",
		URL:    u,
		Body:   io.NopCloser(strings.NewReader(bodyStr)),
		Header: h,
	}

	resp, err := c.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	/*
	// Debug

	fmt.Println(u.String())

	io.Copy(os.Stdout, resp.Body)

	os.Exit(0)

	return "", nil
	*/

	dr := base64.NewDecoder(base64.StdEncoding, resp.Body)

	unparsedd, err := ioutil.ReadAll(dr)

	if err != nil {
		return "", err
	}

	rd := &CrosstabExportResp{}

	err = json.Unmarshal(unparsedd, rd)

	if err != nil {
		return "", err
	}

	if len(rd.VQLCmdResponse.LayoutStatus.ApplicationPresModel.PresentationLayerNotification) < 1 {
		return "", errors.New("missing PresentationLayerNotification")
	}

	return rd.VQLCmdResponse.LayoutStatus.ApplicationPresModel.PresentationLayerNotification[0].PresModelHolder.GenExportFilePresModel.ResultKey, nil
}

func GetCrosstabCSVRequestURL(sessionID string, requestID string) (string, error) {
	u, err := url.Parse("https://visualizedata.ucop.edu/vizql/t/UCSCpublic/w/COVID-19DashboardV2/v/COVID-19Dashboard/tempfile/sessions/?keepfile=yes&attachment=yes")

	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, sessionID + "/")

	q := u.Query()
	q.Set("key", requestID)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// APIHost const
const APIHost = "https://slack.com/api/"

// DialogJSON struct
type DialogJSON struct {
	// Token     string `json:"token"`
	TriggerID string `json:"trigger_id,omitempty"`
	Dialog    Dialog `json:"dialog,omitempty"`
}

// Dialog struct
type Dialog struct {
	CallbackID string      `json:"callback_id,omitempty"`
	Title      string      `json:"title,omitempty"`
	State      interface{} `json:"state,omitempty"`
	Elements   []Elements  `json:"elements,omitempty"`
}

// Elements struct
type Elements struct {
	Type           string    `json:"type,omitempty"`
	Label          string    `json:"label,omitempty"`
	Name           string    `json:"name,omitempty"`
	Value          string    `json:"value,omitempty"`
	Placeholder    string    `json:"placeholder,omitempty"`
	Hint           string    `json:"hint,omitempty"`
	DataSource     string    `json:"data_source,omitempty"`
	Options        []Options `json:"options,omitempty"`
	Optional       bool      `json:"optional,omitempty"`
	MinQueryLength int       `json:"min_query_length,omitempty"`
}

// Options struct
type Options struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// UserGroupJSON struct
type UserGroupJSON struct {
	OK         bool        `json:"ok,omitempty"`
	Error      string      `json:"error,omitempty"`
	Usergroups []UserGroup `json:"usergroups,omitempty"`
	Users      []string    `json:"users,omitempty"`
}

// UserGroup struct
type UserGroup struct {
	ID          string `json:"id,omitempty"`
	TeamID      string `json:"team_id,omitempty"`
	IsUsergroup bool   `json:"is_usergroup,omitempty"`
	IsSubteam   bool   `json:"is_subteam,omitempty"`
	Name        string `json:"name,omitempty"`
	Handle      string `json:"handle,omitempty"`
	IsExternal  bool   `json:"is_external,omitempty"`
	DateCreate  int    `json:"date_create,omitempty"`
	DateUpdate  int    `json:"date_update,omitempty"`
	DateDelete  int    `json:"date_delete,omitempty"`
	UserCount   int    `json:"user_count,omitempty"`
}

// Init func
func Init(oauth, verification string) *Client {
	return &Client{oauth, verification}
}

// Client struct
type Client struct {
	oauth        string
	verification string
}

// DialogResponse struct
type DialogResponse struct {
	OK               bool      `json:"ok,omitempty"`
	Error            string    `json:"error,omitempty"`
	ResponseMetadata *Messages `json:"response_metadata,omitempty"`
}

// Messages struct
type Messages struct {
	Messages []string `json:"messages,omitempty"`
}

// ResponseJSON struct
type ResponseJSON struct {
	Text        string               `json:"text,omitempty"`
	Attachments []ResponseAttachment `json:"attachments,omitempty"`
	Channel     string               `json:"channel,omitempty"`
	User        string               `json:"user,omitempty"`
}

// ResponseAttachment struct
type ResponseAttachment struct {
	Fallback       string                      `json:"fallback,omitempty"`
	Color          string                      `json:"color,omitempty"`
	Pretext        string                      `json:"pretext,omitempty"`
	AuthorName     string                      `json:"author_name,omitempty"`
	AuthorLink     string                      `json:"author_link,omitempty"`
	AuthorIcon     string                      `json:"author_icon,omitempty"`
	Title          string                      `json:"title,omitempty"`
	TitleLink      string                      `json:"title_link,omitempty"`
	Text           string                      `json:"text,omitempty"`
	Fields         []ResponseAttachmentField   `json:"fields"`
	ImageURL       string                      `json:"image_url,omitempty"`
	ThumbURL       string                      `json:"thumb_url,omitempty"`
	Footer         string                      `json:"footer,omitempty"`
	FooterIcon     string                      `json:"footer_icon,omitempty"`
	TS             int                         `json:"ts,omitmepty"`
	Callback       string                      `json:"callback_id,omitmepty"`
	AttachmentType string                      `json:"attachment_type,omitmepty"`
	Actions        []ResponseAttachmentActions `json:"actions"`
}

// ResponseAttachmentActions struct
type ResponseAttachmentActions struct {
	Name    string                              `json:"name"`
	Text    string                              `json:"text"`
	Type    string                              `json:"type"`
	Value   string                              `json:"value"`
	Style   string                              `json:"style"`
	Options *[]ResponseAttachmentActionsOptions `json:"options,omitempty"`
	Confirm *ResponseConfirm                    `json:"confirm,omitempty"`
}

// ResponseConfirm struct
type ResponseConfirm struct {
	Title       string `json:"title,omitempty"`
	Text        string `json:"text,omitempty"`
	TextOK      string `json:"ok_text,omitempty"`
	TextDismiss string `json:"dismiss_text,omitempty"`
}

// ResponseAttachmentActionsOptions struct
type ResponseAttachmentActionsOptions struct {
	Text  string `json:"text,omitempty"`
	Value string `json:"value,omitempty"`
}

// ResponseAttachmentField struct
type ResponseAttachmentField struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
	Short bool   `json:"short,omitempty"`
}

// MessageDeleteJSON struct
type MessageDeleteJSON struct {
	Channel string `json:"channel,omitempty"`
	TS      string `json:"ts,omitempty"`
	// ASUuser bool   `json:"as_user"`
}

// ChatPostMessageJSON struct
type ChatPostMessageJSON struct {
	Channel     string               `json:"channel,omitempty"`
	Text        string               `json:"text,omitempty"`
	AsUser      bool                 `json:"as_user,omitempty"`
	Attachments []ResponseAttachment `json:"attachments,omitempty"`
	IconEmoji   string               `json:"icon_emoji,omitempty"`
	IconURL     string               `json:"icon_url,omitempty"`
	LinkNames   bool                 `json:"link_names,omitempty"`
	ThreadTS    string               `json:"thread_ts,omitempty"`
}

// APIResponseJSON struct
type APIResponseJSON struct {
	OK    bool   `json:"ok,omitempty"`
	Error string `json:"error,omitempty"`
}

// DialogOpen func
func (c *Client) DialogOpen(triggerID string, dialog Dialog) error {
	// json, err := json.MarshalIndent(DialogJSON{triggerID, dialog}, "", "    ")
	out, err := json.Marshal(DialogJSON{triggerID, dialog})
	if err != nil {
		return err
	}

	log.Printf("DialogOpen: %+v", string(out))

	body, err := c.method("dialog.open", out)

	if err != nil {
		log.Print(string(body))
		return err
	}

	res := DialogResponse{}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	log.Printf("DialogResponse: %+v", res)

	if !res.OK {
		return fmt.Errorf(res.Error)
	}

	return nil
}

// UserGroupList func
func (c *Client) UserGroupList() ([]UserGroup, error) {
	body, err := c.method("usergroups.list", nil)

	res := UserGroupJSON{}
	if err = json.Unmarshal(body, &res); err != nil {
		return res.Usergroups, err
	}

	if !res.OK {
		return res.Usergroups, fmt.Errorf(res.Error)
	}

	return res.Usergroups, nil
}

// MessageDelete func
func (c *Client) MessageDelete(chatID, messageID string) error {
	res := APIResponseJSON{}

	out, err := json.Marshal(MessageDeleteJSON{chatID, messageID})
	if err != nil {
		return err
	}

	log.Print(string(out))
	body, err := c.method("chat.delete", out)
	log.Print(string(body))

	if err = json.Unmarshal(body, &res); err != nil {
		return err
	}

	if !res.OK {
		return fmt.Errorf(res.Error)
	}

	return nil
}

// Response func
func (c *Client) Response(url string, response ResponseJSON) ([]byte, error) {
	OutputResult, err := json.Marshal(response)
	fmt.Print("\n-----------\n")
	log.Print("Send json: ", string(OutputResult))

	if err != nil {
		OutputResult = []byte(fmt.Sprintf("output marshal error: %s", err.Error()))
	}

	res, err := c.send(url, OutputResult)

	log.Print(fmt.Sprintf("Response: %+v, %+v", string(res), err))
	fmt.Print("-----------\n")

	return res, err
}

// ResponseDelay func
func (c *Client) ResponseDelay(response ResponseJSON) ([]byte, error) {
	OutputResult, err := json.Marshal(response)
	fmt.Print("\n-----------\n")
	log.Print("Send json: ", string(OutputResult))

	if err != nil {
		OutputResult = []byte(fmt.Sprintf("output marshal error: %s", err.Error()))
	}

	res, err := c.method("chat.postEphemeral", OutputResult)

	log.Print(fmt.Sprintf("ResponseDelay: %+v, %+v", string(res), err))
	fmt.Print("-----------\n")

	return res, err
}

// ChatPost func
func (c *Client) ChatPost(response ChatPostMessageJSON) ([]byte, error) {
	OutputResult, err := json.Marshal(response)
	fmt.Print("\n-----------\n")
	log.Print("Send json: ", string(OutputResult))

	if err != nil {
		OutputResult = []byte(fmt.Sprintf("output marshal error: %s", err.Error()))
	}

	res, err := c.method("chat.postMessage", OutputResult)

	log.Print(fmt.Sprintf("ChatPost: %+v, %+v", string(res), err))
	fmt.Print("-----------\n")

	return res, err
}

// UserGroupFindUser func
func (c *Client) UserGroupFindUser(usergroup, user string) (bool, error) {
	res, err := c.UserGroupUsers(usergroup)
	if err != nil {
		return false, err
	}

	users := map[string]bool{}
	for _, j := range res {
		users[j] = true
	}

	return users[user], nil
}

// UserGroupUsers func
func (c *Client) UserGroupUsers(usergroup string) ([]string, error) {
	res := UserGroupJSON{}

	fmt.Print("\n-----------\n")
	body, err := c.method(fmt.Sprintf("usergroups.users.list?usergroup=%s", usergroup), nil)

	log.Print(fmt.Sprintf("UserGroupUsers: %+v, %+v", string(body), err))
	fmt.Print("-----------\n")

	if err = json.Unmarshal(body, &res); err != nil {
		return res.Users, err
	}

	if !res.OK {
		return res.Users, fmt.Errorf(res.Error)
	}

	return res.Users, err
}

func (c *Client) method(method string, json []byte) ([]byte, error) {
	return c.send(APIHost+"/"+method, json)
}

func (c *Client) send(url string, json []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Authorization", "Bearer "+c.oauth)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Printf("Reponse: %+v", resp)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad status code (" + strconv.Itoa(resp.StatusCode) + ")")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

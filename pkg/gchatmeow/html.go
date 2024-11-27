package gchatmeow

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/pblite"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto/gchatprotoweb"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/types"

	"golang.org/x/net/html"
)

var jsObjectRe = regexp.MustCompile(`(?m)(\s*{|\s*,\s*)\s*([a-zA-Z0-9_]+)\s*:`)
var jsValueRe = regexp.MustCompile(`:\s*'([^']*)'`)

func (c *Client) parseInitialMessagesHTML(tags []ScriptTag) (*types.InitialConfigData, error) {
	var wizGlobalData types.MessagesPageConfig
	var getSelfData gchatprotoweb.DynamiteGetSelf
	var groupListData gchatprotoweb.DynamiteGetGroupList
	var err error
	for _, tag := range tags {
		if dataId, ok := tag.Attributes["data-id"]; ok {
			switch dataId {
			// wiz global data
			case "_gd":
				jsonContent := strings.Replace(strings.TrimSuffix(strings.TrimRight(tag.Content, "\n "), ";"), "window.WIZ_global_data = ", "", -1)
				err = json.Unmarshal([]byte(jsonContent), &wizGlobalData)
				if err != nil {
					return nil, err
				}
				if wizGlobalData.UserID == "" {
					return nil, fmt.Errorf("failed to parse authenticated user info from main messaging page, ensure your cookies are valid")
				}
			default:
				c.Logger.Warn().Str("data-id", dataId).Str("content", tag.Content).Msg("Found unknown/unhandled data-id attribute key in html response")
			}
		} else if strings.Contains(tag.Content, "AF_initDataCallback") && strings.Contains(tag.Content, "key") {
			jsonObjectData := parseAfInitDataObject(tag.Content)
			switch jsonObjectData.Key {
			case types.DynamiteGetSelf:
				err = pblite.UnmarshalSlice(jsonObjectData.Data, &getSelfData)
				if err != nil {
					return nil, err
				}
			case types.DynamiteGetGroupList:
				err = pblite.UnmarshalSlice(jsonObjectData.Data, &groupListData)
				if err != nil {
					return nil, err
				}
			default:
				break
			}
		}
	}
	return &types.InitialConfigData{
		PageConfig:  &wizGlobalData,
		CurrentUser: &getSelfData,
		GroupList:   &groupListData,
	}, nil
}

func parseAfInitDataObject(content string) *types.AFConfigData {
	jsonObjectString := PreprocessJSObject(strings.TrimSuffix(strings.TrimRight(strings.Replace(content, "AF_initDataCallback(", "", -1), "\n "), ");"))
	var data types.AFConfigData
	err := json.Unmarshal([]byte(jsonObjectString), &data)
	if err != nil {
		log.Fatal(err)
	}
	return &data
}

func PreprocessJSObject(s string) string {
	keysFixed := jsObjectRe.ReplaceAllString(s, "$1 \"$2\":")
	return jsValueRe.ReplaceAllString(keysFixed, `: "$1"`)
}

type NodeProcessor func(*html.Node) interface{}

func processNode(n *html.Node, tag string, processor NodeProcessor) interface{} {
	if n.Type == html.ElementNode && n.Data == tag {
		return processor(n)
	}
	return nil
}

func recursiveSearch(n *html.Node, tag string, processor NodeProcessor) []interface{} {
	var result []interface{}

	if item := processNode(n, tag, processor); item != nil {
		result = append(result, item)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result = append(result, recursiveSearch(c, tag, processor)...)
	}

	return result
}

type ScriptTag struct {
	Attributes map[string]string
	Content    string
}

func findTags(tag string, processor NodeProcessor, n *html.Node) []interface{} {
	return recursiveSearch(n, tag, processor)
}

func findScriptTags(n *html.Node) []ScriptTag {
	processor := func(n *html.Node) interface{} {
		attributes := make(map[string]string)
		for _, a := range n.Attr {
			attributes[a.Key] = a.Val
		}
		content := ""
		if n.FirstChild != nil {
			content = n.FirstChild.Data
		}
		return ScriptTag{Attributes: attributes, Content: content}
	}

	tags := findTags("script", processor, n)
	scriptTags := make([]ScriptTag, len(tags))
	for i, t := range tags {
		scriptTags[i] = t.(ScriptTag)
	}
	return scriptTags
}

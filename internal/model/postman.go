package model

import "encoding/json"

const PostmanSchemaV210 = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"

// PostmanCollection represents a complete Postman Collection v2.1.0.
type PostmanCollection struct {
	Info     PostmanInfo   `json:"info"`
	Item     []PostmanItem `json:"item"`
	Variable []PostmanVar  `json:"variable,omitempty"`
}

// PostmanInfo contains the collection metadata.
type PostmanInfo struct {
	Name        string `json:"name"`
	PostmanID   string `json:"_postman_id"`
	Description string `json:"description"`
	Schema      string `json:"schema"`
}

// PostmanItem can be a folder (has Item) or a request (has Request).
type PostmanItem struct {
	Name    string         `json:"name"`
	Item    []PostmanItem  `json:"item,omitempty"`
	Request *PostmanReq    `json:"request,omitempty"`
	Auth    *PostmanAuth   `json:"auth,omitempty"`
}

// IsFolder returns true if the item is a folder (contains sub-items, no request).
func (i PostmanItem) IsFolder() bool {
	return i.Request == nil
}

// PostmanReq represents an HTTP request in Postman format.
type PostmanReq struct {
	Method string          `json:"method"`
	URL    PostmanURL      `json:"url"`
	Header []PostmanHeader `json:"header,omitempty"`
	Body   *PostmanBody    `json:"body,omitempty"`
	Auth   *PostmanAuth    `json:"auth,omitempty"`
}

// PostmanURL represents a URL that can be either a string or an object with "raw" field.
type PostmanURL struct {
	Raw string `json:"raw"`
}

// MarshalJSON serializes the URL as a plain string for maximum Postman compatibility.
func (u PostmanURL) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.Raw)
}

// UnmarshalJSON handles both string and object URL formats from Postman.
func (u *PostmanURL) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err == nil {
		u.Raw = raw
		return nil
	}

	var obj struct {
		Raw string `json:"raw"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	u.Raw = obj.Raw
	return nil
}

// PostmanHeader represents an HTTP header.
type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PostmanBody represents the request body.
type PostmanBody struct {
	Mode     string           `json:"mode"`
	Raw      string           `json:"raw,omitempty"`
	FormData []PostmanFormData `json:"formdata,omitempty"`
}

// PostmanFormData represents a form field.
type PostmanFormData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PostmanAuth represents authentication configuration.
type PostmanAuth struct {
	Type   string     `json:"type"`
	Bearer []PostmanKV `json:"bearer,omitempty"`
}

// PostmanKV is a generic key-value pair.
type PostmanKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PostmanVar represents a collection variable.
type PostmanVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

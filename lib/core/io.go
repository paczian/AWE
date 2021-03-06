package core

import (
	"errors"
	"fmt"
	"github.com/MG-RAST/AWE/lib/logger"
	"github.com/MG-RAST/AWE/lib/shock"
	"github.com/MG-RAST/golib/go-uuid/uuid"
	"net/url"
	"strings"
)

type IO struct {
	FileName      string                   `bson:"filename" json:"filename" mapstructure:"filename"`
	Name          string                   `bson:"name" json:"name" mapstructure:"name"`  // specifies abstract name of output as defined by the app
	AppPosition   int                      `bson:"appposition" json:"-" mapstructure:"-"` // specifies position in app output array
	Directory     string                   `bson:"directory" json:"directory" mapstructure:"directory"`
	Host          string                   `bson:"host" json:"host" mapstructure:"host"`
	Node          string                   `bson:"node" json:"node" mapstructure:"node"`
	Url           string                   `bson:"url"  json:"url" mapstructure:"url"` // can be shock or any other url
	Size          int64                    `bson:"size" json:"size" mapstructure:"size"`
	MD5           string                   `bson:"md5" json:"-" mapstructure:"-"`
	Cache         bool                     `bson:"cache" json:"cache" mapstructure:"cache"` // indicates that this files is "predata"" that needs to be cached
	Origin        string                   `bson:"origin" json:"origin" mapstructure:"origin"`
	Path          string                   `bson:"-" json:"-" mapstructure:"-"`
	Optional      bool                     `bson:"optional" json:"-" mapstructure:"-"`
	Nonzero       bool                     `bson:"nonzero"  json:"nonzero" mapstructure:"nonzero"`
	DataToken     string                   `bson:"datatoken"  json:"-" mapstructure:"-"`
	Intermediate  bool                     `bson:"Intermediate"  json:"-" mapstructure:"-"`
	Temporary     bool                     `bson:"temporary"  json:"temporary" mapstructure:"temporary"`
	ShockFilename string                   `bson:"shockfilename" json:"shockfilename" mapstructure:"shockfilename"`
	ShockIndex    string                   `bson:"shockindex" json:"shockindex" mapstructure:"shockindex"` // on input it indicates that Shock node has to be indexed by AWE server
	AttrFile      string                   `bson:"attrfile" json:"attrfile" mapstructure:"attrfile"`
	NoFile        bool                     `bson:"nofile" json:"nofile" mapstructure:"nofile"`
	Delete        bool                     `bson:"delete" json:"delete" mapstructure:"delete"` // speficies that this is a temorary node, to be deleted from shock on job completion
	Type          string                   `bson:"type" json:"type" mapstructure:"type"`
	NodeAttr      map[string]interface{}   `bson:"nodeattr" json:"nodeattr" mapstructure:"nodeattr"` // specifies attribute data to be stored in shock node (output only)
	FormOptions   map[string]string        `bson:"formoptions" json:"formoptions" mapstructure:"formoptions"`
	Uncompress    string                   `bson:"uncompress" json:"uncompress" mapstructure:"uncompress"` // tells AWE client to uncompress this file, e.g. "gzip"
	Indexes       map[string]shock.IdxInfo `bson:"-" json:"-" mapstructure:"-"`                            // copy of shock node.Indexes
}

type PartInfo struct {
	Input         string `bson:"input" json:"input" mapstructure:"input"`
	Index         string `bson:"index" json:"index" mapstructure:"index"`
	TotalIndex    int    `bson:"totalindex" json:"totalindex" mapstructure:"totalindex"`
	MaxPartSizeMB int    `bson:"maxpartsize_mb" json:"maxpartsize_mb" mapstructure:"maxpartsize_mb"`
	Options       string `bson:"options" json:"-" mapstructure:"-"`
}

// Deprecated JobDep struct uses deprecated TaskDep struct which uses the deprecated IOmap.  Maintained for backwards compatibility.
// Jobs that cannot be parsed into the Job struct, but can be parsed into the JobDep struct will be translated to the new Job struct.
// (=deprecated=)
type IOmap map[string]*IO // [filename]attributes

// (=deprecated=)
func NewIOmap() IOmap {
	return IOmap{}
}

// (=deprecated=)
func (i IOmap) Add(name string, host string, node string, md5 string, cache bool) {
	i[name] = &IO{FileName: name, Host: host, Node: node, MD5: md5, Cache: cache}
	return
}

// (=deprecated=)
func (i IOmap) Has(name string) bool {
	if _, has := i[name]; has {
		return true
	}
	return false
}

// (=deprecated=)
func (i IOmap) Find(name string) *IO {
	if val, has := i[name]; has {
		return val
	}
	return nil
}

func NewIO() *IO {
	return &IO{}
}

func (io *IO) Url2Shock() (err error) {
	u, _ := url.Parse(io.Url)
	if (u.Scheme == "") || (u.Host == "") || (u.Path == "") {
		err = fmt.Errorf("Not a valid url: %s", io.Url)
		return
	}
	// get shock info from url
	if (io.Host == "") || (io.Node == "") || (io.Node == "-") {
		trimPath := strings.Trim(u.Path, "/")
		cleanUuid := strings.Trim(strings.TrimPrefix(trimPath, "node"), "/")
		// appears to be a shock url
		if (cleanUuid != trimPath) && (uuid.Parse(cleanUuid) != nil) {
			io.Host = u.Scheme + "://" + u.Host
			io.Node = cleanUuid
		}
	}
	return
}

func (io *IO) DataUrl() (dataurl string, err error) {
	if io.Url != "" {
		// parse and test url
		err = io.Url2Shock()
		if err != nil {
			return
		}
		return io.Url, nil
	} else if (io.Host != "") && (io.Node != "") && (io.Node != "-") {
		io.Url = fmt.Sprintf("%s/node/%s%s", io.Host, io.Node, shock.DATA_SUFFIX)
		return io.Url, nil
	} else {
		// empty IO is valid
		return "", nil
	}
}

func (io *IO) TotalUnits(indextype string) (count int, err error) {
	count, err = io.getIndexUnits(indextype)
	return
}

func (io *IO) HasFile() bool {
	// set io.Size and io.MD5
	shocknode, err := io.GetShockNode()
	if err != nil {
		logger.Error(fmt.Sprintf("HasFile error: %s, node: %s", err.Error(), io.Node))
		return false
	}
	io.Size = shocknode.File.Size
	if md5, ok := shocknode.File.Checksum["md5"]; ok {
		io.MD5 = md5
	}
	// both can not be empty
	if (io.Size == 0) && (io.MD5 == "") {
		return false
	}
	return true
}

func (io *IO) GetFileSize() (size int64, modified bool, err error) {
	modified = false
	if io.Size > 0 {
		size = io.Size
		return
	}
	shocknode, err := io.GetShockNode()
	if err != nil {
		err = fmt.Errorf("GetFileSize error: %s, node: %s", err.Error(), io.Node)
		return
	}
	if (shocknode.File.Size == 0) && shocknode.File.CreatedOn.IsZero() {
		msg := "Node has no file"
		if (shocknode.Type == "parts") && (shocknode.Parts != nil) {
			msg += fmt.Sprintf(", %d of %d parts completed", shocknode.Parts.Length, shocknode.Parts.Count)
		}
		err = fmt.Errorf("GetFileSize error: %s, node: %s", msg, io.Node)
		return
	}
	size = shocknode.File.Size
	if size != io.Size {
		io.Size = size
		modified = true
	}
	return
}

func (io *IO) GetIndexInfo(indextype string) (idxInfo shock.IdxInfo, hasIndex bool, err error) {
	if idxInfo, hasIndex = io.Indexes[indextype]; hasIndex {
		return
	}
	// missing, update io.Indexes from shock
	_, err = io.GetShockNode()
	if err != nil {
		return
	}
	idxInfo, hasIndex = io.Indexes[indextype]
	return
}

func (io *IO) GetShockNode() (node *shock.ShockNode, err error) {
	if io.Host == "" {
		err = errors.New("empty shock host")
		return
	}
	if io.Node == "-" {
		err = errors.New("empty node id")
		return
	}
	node, err = shock.ShockGet(io.Host, io.Node, io.DataToken)
	if err != nil {
		return
	}
	// always update indexinfo with shock GET
	io.Indexes = node.Indexes
	return
}

func (io *IO) getIndexUnits(indextype string) (totalunits int, err error) {
	idxInfo, hasIndex, err := io.GetIndexInfo(indextype)
	if err != nil {
		return
	}
	if !hasIndex {
		err = fmt.Errorf("getIndexUnits error: shock node %s has no indextype %s", io.Node, indextype)
		return
	}
	if idxInfo.TotalUnits > 0 {
		totalunits = int(idxInfo.TotalUnits)
		return
	}
	err = fmt.Errorf("getIndexUnits error: invalid totalunits for shock node %s, indextype %s", io.Node, indextype)
	return
}

func (io *IO) DeleteNode() (err error) {
	err = shock.ShockDelete(io.Host, io.Node, io.DataToken)
	return
}

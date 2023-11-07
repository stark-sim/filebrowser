package bd

type FileMetasArg struct {
	Fsids []uint64 `json:"fsids"`
	Path  string   `json:"path"` //查询共享目录或专属空间内文件时需要
}

func NewFileMetasArg(fsid []uint64, path string) *FileMetasArg {
	s := new(FileMetasArg)
	s.Fsids = fsid
	s.Path = path
	return s
}

type ListInfo struct {
	Size        uint64            `json:"Size"`
	Path        string            `json:"path"`
	Isdir       int               `json:"isdir"`
	ServerCtime uint64            `json:"server_ctime"`
	ServerMtime uint64            `json:"server_mtime"`
	Fsid        uint64            `json:"fs_id"`
	OperId      int               `json:"oper_id"`
	Md5         string            `json:"md5"`
	Filename    string            `json:"filename"`
	Category    int               `json:"category"`
	Dlink       string            `json:"dlink"` // 文件才返回dlink
	Duration    int               `json:"duration"`
	Thumbs      map[string]string `json:"thumbs"`
	Height      int               `json:"height"`
	Width       int               `json:"width"`
	DateTaken   int               `json:"date_taken"`
}

type FileMetasReturn struct {
	Errno     int                    `json:"errno"`
	Errmsg    string                 `json:"errmsg"`
	RequestID string                 `json:"request_id"`
	Names     map[string]interface{} `json:"names"`
	List      []ListInfo             `json:"list"`
}

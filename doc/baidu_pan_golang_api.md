# 百度网盘开放平台 Golang API 接入笔记

本文件整理自百度网盘开放平台文档，供后续 session 和本地开发直接阅读。当前重点关注 Go 后端接入：授权、文件列表、文件管理、上传、下载、分享等基础能力。

最后核对时间：2026-04-11

## 官方文档入口

- 平台简介：https://pan.baidu.com/union/doc/nksg0sbfs
- 授权介绍：https://pan.baidu.com/union/doc/ol0rsap9s
- 获取用户信息：https://pan.baidu.com/union/doc/pksg0s9ns
- 获取网盘容量信息：https://pan.baidu.com/union/doc/Cksg0s9ic
- 获取文件列表：https://pan.baidu.com/union/doc/nksg0sat9
- 递归获取文件列表：https://pan.baidu.com/union/doc/Zksg0sb73
- 查询文件信息：https://pan.baidu.com/union/doc/Fksg0sbcm
- 搜索文件：https://pan.baidu.com/union/doc/zksg0sb9z
- 语义搜索：https://pan.baidu.com/union/doc/1mgk93xgm
- 管理文件：https://pan.baidu.com/union/doc/mksg0s9l4
- 上传能力说明：https://pan.baidu.com/union/doc/3ksg0s9ye
- 获取上传域名：https://pan.baidu.com/union/doc/Mlvw5hfnr
- 预上传：https://pan.baidu.com/union/doc/3ksg0s9r7
- 分片上传：https://pan.baidu.com/union/doc/nksg0s9vi
- 创建文件/文件夹：https://pan.baidu.com/union/doc/rksg0sa17
- 下载：https://pan.baidu.com/union/doc/pkuo3snyp
- 文件分享服务：https://pan.baidu.com/union/doc/Dlagruvfh

## 接入前置条件

使用 API 前需要完成：

- 注册并登录百度账号。
- 完成开发者实名认证。
- 在开放平台控制台创建应用，获得 AppID、AppKey、SecretKey、SignKey 等。
- 接入 OAuth2 授权，获得用户 `access_token`。

授权注意点：

- `access_token` 有效期 30 天。
- `refresh_token` 使用一次后失效。
- 刷新失败后旧 `refresh_token` 也会失效，需要重新发起用户授权。
- 所有文件 API 都需要用户授权后的 `access_token`。

Go 实现建议：

- 用一个独立 client 封装百度网盘 HTTP 请求。
- 所有接口方法都接收 `context.Context`。
- `access_token` 由业务层或 token store 提供，不要散落在接口内部硬编码。
- 远程请求的重试次数、请求间隔、失败重试间隔从配置读取。
- 每两个远程请求之间需要有配置化 sleep；本地 MD5、DTO 组装、文件切片等本地操作之间不需要 sleep。

## Client 结构建议

```go
type Client struct {
	httpClient *http.Client
	baseURL    string
	pcsURL     string
	appID      int64
	userAgent  string
	retry      RetryConfig
}

type RetryConfig struct {
	MaxRetries          int
	RequestInterval     time.Duration
	FailureRetryInterval time.Duration
}
```

建议默认端点：

```go
const (
	panHost = "https://pan.baidu.com"
	pcsHost = "https://d.pcs.baidu.com"
	defaultUserAgent = "pan.baidu.com"
)
```

统一响应可以先定义基础错误字段：

```go
type BaseResponse struct {
	Errno     int64  `json:"errno"`
	RequestID string `json:"request_id"`
	ShowMsg   string `json:"show_msg"`
}
```

注意：不同接口错误字段不完全一致，有的叫 `errno`，有的叫 `error_code`。解析时要按接口分别定义 DTO。

## 用户与容量

### 获取用户信息

接口：

```text
GET /rest/2.0/xpan/nas?method=uinfo
Host: pan.baidu.com
```

参数：

- `method=uinfo`
- `access_token`
- `vip_version=v2`，可选，传 `v2` 时返回实际身份信息。

返回：

- `baidu_name`
- `netdisk_name`
- `avatar_url`
- `vip_type`
- `uk`

用途：

- 展示当前绑定的百度网盘账号。
- 判断会员类型，辅助上传大小限制提示。

### 获取网盘容量

接口：

```text
GET /api/quota
Host: pan.baidu.com
```

参数：

- `access_token`
- `checkfree`，是否检查免费容量。
- `checkexpire`，是否检查容量过期信息。

返回：

- `total`：总容量，单位 B。
- `used`：已用容量，单位 B。
- `free`：免费容量，单位 B。
- `expire`：7 天内是否有容量到期。

## 文件列表与查询

### 列出目录

接口：

```text
GET /rest/2.0/xpan/file?method=list
Host: pan.baidu.com
```

参数：

- `method=list`
- `access_token`
- `dir`：目录绝对路径，默认 `/`，中文路径需要 URL encode。
- `order`：`name`、`time`、`size`。
- `desc`：`0` 升序，`1` 降序。
- `start`：起始位置，从 0 开始。
- `limit`：数量，默认 1000，建议最大不超过 1000。
- `web=1`：返回缩略图和 `dir_empty`。
- `folder=1`：只返回文件夹。
- `showempty=1`：返回 `dir_empty`。

返回核心字段：

- `fs_id`
- `path`
- `server_filename`
- `size`
- `server_mtime`
- `server_ctime`
- `local_mtime`
- `local_ctime`
- `isdir`
- `category`

Go 方法建议：

```go
func (c *Client) ListFiles(ctx context.Context, token string, req ListFilesRequest) (*ListFilesResponse, error)
```

### 递归列出目录

接口：

```text
GET /rest/2.0/xpan/multimedia?method=listall
Host: pan.baidu.com
```

参数：

- `method=listall`
- `access_token`
- `path`：目录绝对路径，必须 `/` 开头。
- `recursion=1`：递归。
- `order`：`time`、`name`、`size`。
- `desc`：`0` 升序，`1` 降序。
- `start`：查询起点；有下一页时使用响应的 `cursor`。
- `limit`：建议最大 1000。
- `ctime`：只返回上传时间大于该时间的文件。
- `mtime`：只返回修改时间大于该时间的文件。
- `web=1`：返回缩略图。

返回：

- `has_more`
- `cursor`
- `list`

用途：

- 扫描某个目录下的全部文件。
- 做同步任务、目录索引、增量扫描。

### 查询文件信息

接口：

```text
GET /rest/2.0/xpan/multimedia?method=filemetas
Host: pan.baidu.com
```

参数：

- `method=filemetas`
- `access_token`
- `fsids`：文件 ID 数组，最多 100 个。
- `dlink=1`：返回下载地址。
- `thumb=1`：返回缩略图地址。
- `extra=1`：图片返回拍摄时间、原图分辨率等。
- `needmedia=1`：视频返回时长。
- `detail=1`：视频返回宽高等媒体信息。

用途：

- 通过 `fs_id` 批量获取文件详情。
- 下载前获取 `dlink`。
- 展示图片缩略图、视频时长等。

Go 方法建议：

```go
func (c *Client) FileMetas(ctx context.Context, token string, fsids []uint64, opt FileMetaOption) (*FileMetasResponse, error)
```

## 搜索与分类

### 关键词搜索

接口：

```text
GET /rest/2.0/xpan/file?method=search
Host: pan.baidu.com
```

参数：

- `method=search`
- `access_token`
- `key`：搜索关键字，最大 30 字符。
- `dir`：搜索目录，默认根目录。
- `category`：文件类型。
- `num`：默认 500，官方说明不能修改。
- `recursion=1`：递归搜索。
- `web=1`：返回缩略图。

文件类型：

- `1` 视频
- `2` 音频
- `3` 图片
- `4` 文档
- `5` 应用
- `6` 其他
- `7` 种子

### 语义搜索

接口：

```text
POST /xpan/unisearch
Host: pan.baidu.com
```

能力：

- 文件名关键词搜索。
- 按时间、类型、后缀、大小、时长、路径搜索。
- 图片 OCR 搜索、图片语义搜索。
- 文档内容搜索、文档语义搜索。
- 视频、音频 ASR 搜索和语义搜索。

参数：

- `access_token`
- `scene=mcpserver`
- `query`
- `dir`：路径数组。
- `category`：类型数组。
- `num`
- `stream`
- `search_type`：`0` 简单搜索，`1` 语义搜索，`2` 自动。
- `sources`：召回来源。

此接口较新，接入前建议再次核对官方文档。

### 分类文件列表

接口：

```text
GET /rest/2.0/xpan/multimedia?method=categorylist
Host: pan.baidu.com
```

参数：

- `category`：可传多个，英文逗号分隔，如 `3,4`。
- `show_dir`：是否展示文件夹。
- `parent_path`
- `recursion`
- `ext`：文件后缀，多个英文逗号分隔。
- `start`
- `limit`：最大 1000。
- `order`
- `desc`

用途：

- 获取图片、文档、音频、视频等分类列表。
- 按扩展名过滤，如 `txt,epub`。

## 文件管理

接口：

```text
POST /rest/2.0/xpan/file?method=filemanager
Host: pan.baidu.com
```

URL 参数：

- `method=filemanager`
- `access_token`
- `opera`：`copy`、`move`、`rename`、`delete`。

Body 参数：

- `async`：
  - `0` 同步
  - `1` 自适应
  - `2` 异步
- `filelist`：待操作文件 JSON。
- `ondup`：重名策略。

重名策略：

- `fail`：失败，默认。
- `newcopy`：自动重命名。
- `overwrite`：覆盖。
- `skip`：跳过。

### 复制或移动

`filelist` 示例：

```json
[
  {
    "path": "/test/123456.docx",
    "dest": "/test/abc",
    "newname": "11223.docx"
  }
]
```

Go 方法建议：

```go
func (c *Client) MoveFile(ctx context.Context, token string, srcPath, destDir, newName string, ondup string) (*FileManagerResponse, error)
func (c *Client) CopyFile(ctx context.Context, token string, srcPath, destDir, newName string, ondup string) (*FileManagerResponse, error)
```

### 重命名

`filelist` 示例：

```json
[
  {
    "path": "/test/123456.docx",
    "newname": "123.docx"
  }
]
```

Go 方法建议：

```go
func (c *Client) RenameFile(ctx context.Context, token string, path, newName string, ondup string) (*FileManagerResponse, error)
```

### 删除

`filelist` 示例：

```json
[
  "/test/123456.docx"
]
```

注意：

- `delete` 操作没有 `ondup`。
- 失败后可以再次发送请求。

Go 方法建议：

```go
func (c *Client) DeleteFiles(ctx context.Context, token string, paths []string) (*FileManagerResponse, error)
```

## 创建文件夹

接口：

```text
POST /rest/2.0/xpan/file?method=create
Host: pan.baidu.com
```

Body 参数：

- `path`：文件夹绝对路径，需要 URL encode。
- `isdir=1`
- `rtype`：
  - `0` 不重命名，冲突返回错误。
  - `1` 冲突时重命名。
- `local_ctime`
- `local_mtime`
- `mode`

示例：

```text
path=/apps/AppName/mydir&isdir=1&rtype=0
```

返回：

- `fs_id`
- `category=6`
- `path`
- `ctime`
- `mtime`
- `isdir=1`

Go 方法建议：

```go
func (c *Client) CreateFolder(ctx context.Context, token string, path string, rtype int) (*CreateFileResponse, error)
```

## 上传

官方推荐主流程：预上传、分片上传、创建文件。

### 上传限制

- 上传路径必须在 `/apps/{appname}` 下。
- 普通用户单文件最大 4GB。
- 会员用户单文件最大 10GB。
- 超级会员单文件最大 20GB。
- 分片数量不得超过 1024。
- 小于等于 4MB 的文件也可以走分片接口，只上传一个分片。
- 单步上传上限 2GB，但官方推荐主流接口使用分片上传。

### 本地准备

上传前在本地计算：

- 文件大小 `size`。
- 文件完整 MD5，即 `content-md5`。
- 文件前 256KB 的 MD5，即 `slice-md5`。
- 按分片计算 MD5，组成 `block_list`。

分片规则：

- 文件小于等于 4MB：`block_list` 只有一个元素，值为整个文件 MD5。
- 文件大于 4MB：按 4MB 切片，每片计算 32 位小写 MD5，组成 JSON 字符串数组。

Go 方法建议：

```go
type LocalPart struct {
	Seq  int
	MD5  string
	Size int64
	From int64
	To   int64
}

func BuildUploadParts(filePath string, partSize int64) (size int64, contentMD5 string, sliceMD5 string, parts []LocalPart, err error)
```

### 预上传

接口：

```text
POST /rest/2.0/xpan/file?method=precreate&access_token={access_token}
Host: pan.baidu.com
```

Body 参数：

- `path`
- `size`
- `isdir=0`
- `block_list`
- `autoinit=1`
- `rtype`
- `content-md5`
- `slice-md5`
- `local_ctime`
- `local_mtime`

`rtype`：

- `1` 路径冲突时重命名。
- `2` 路径冲突且 `block_list` 不同时重命名。
- `3` 覆盖同名文件。

返回：

- `uploadid`
- `block_list`：需要上传的分片序号列表，从 0 开始。

Go 方法建议：

```go
func (c *Client) Precreate(ctx context.Context, token string, req PrecreateRequest) (*PrecreateResponse, error)
```

### 获取上传域名

接口：

```text
GET /rest/2.0/pcs/file?method=locateupload
Host: d.pcs.baidu.com
```

参数：

- `method=locateupload`
- `appid=250528`
- `access_token`
- `path`
- `uploadid`
- `upload_version=2.0`

返回：

- `servers`：选择其中 `https` 协议的域名，例如 `https://c3.pcs.baidu.com`。

Go 方法建议：

```go
func (c *Client) LocateUpload(ctx context.Context, token string, path string, uploadID string) (uploadHost string, err error)
```

### 分片上传

接口：

```text
POST {uploadHost}/rest/2.0/pcs/superfile2?method=upload&access_token={access_token}&type=tmpfile&path={path}&uploadid={uploadid}&partseq={partseq}
```

Body：

- `multipart/form-data`
- 表单字段 `file`：分片内容。

参数：

- `type=tmpfile`
- `path`：必须与预上传一致。
- `uploadid`：预上传返回值。
- `partseq`：分片序号，从 0 开始。

返回：

- `errno`
- `md5`：云端分片 MD5。

注意：

- 创建文件阶段需要使用分片上传返回的 `md5`。
- `block_list` 顺序必须与 `partseq` 顺序一致。
- 普通用户分片大小固定为 4MB。
- 普通会员分片最大 16MB。
- 超级会员分片最大 32MB。
- 项目实现可以统一用 4MB 分片，兼容性最简单。

Go 方法建议：

```go
func (c *Client) UploadPart(ctx context.Context, token string, uploadHost string, req UploadPartRequest, r io.Reader) (*UploadPartResponse, error)
```

### 创建文件

接口：

```text
POST /rest/2.0/xpan/file?method=create&access_token={access_token}
Host: pan.baidu.com
```

Body 参数：

- `path`
- `size`
- `isdir=0`
- `block_list`：分片上传返回的 md5 数组，按序排列。
- `uploadid`
- `rtype`
- `local_ctime`
- `local_mtime`

返回：

- `fs_id`
- `md5`
- `server_filename`
- `category`
- `path`
- `size`
- `ctime`
- `mtime`
- `isdir`

Go 方法建议：

```go
func (c *Client) CreateFile(ctx context.Context, token string, req CreateFileRequest) (*CreateFileResponse, error)
```

### 上传总入口

建议对业务层只暴露一个入口：

```go
func (c *Client) UploadFile(ctx context.Context, token string, localPath string, remotePath string, opt UploadOption) (*CreateFileResponse, error)
```

内部流程：

```text
1. 本地计算 size、contentMD5、sliceMD5、分片 MD5。
2. 调用 Precreate。
3. 调用 LocateUpload 获取上传域名。
4. 按 Precreate 返回的 block_list 上传所需分片。
5. 收集 UploadPart 返回的 md5。
6. 调用 CreateFile 合并文件。
```

## 下载

官方下载流程：

```text
1. 获取文件列表，拿到 fs_id。
2. 调用 filemetas，并传 dlink=1，拿到 dlink。
3. 使用 dlink + access_token 下载文件。
```

下载请求：

```text
GET {dlink}&access_token={access_token}
Host: d.pcs.baidu.com
Header: User-Agent: pan.baidu.com
```

注意：

- `dlink` 有效期 8 小时。
- 必须拼接 `access_token`。
- 必须设置 `User-Agent: pan.baidu.com`。
- `dlink` 存在 302 跳转。
- 超过 50MB 的文件不允许浏览器直接下载，需要原生软件或 App 下载。
- 支持 `Range` 断点续传。

Range 示例：

```text
Range: bytes=0-499
Range: bytes=500-999
Range: bytes=-500
Range: bytes=500-
```

Go 方法建议：

```go
func (c *Client) GetDownloadLink(ctx context.Context, token string, fsid uint64) (string, error)
func (c *Client) DownloadFile(ctx context.Context, token string, dlink string, dst io.Writer, opt DownloadOption) error
```

## 文件分享服务

文件分享服务是付费能力，且属于企业开发者专属权益。接入前需要确认账号已完成企业认证并已购买服务。

能力：

- 创建分享链接。
- 设置分享有效期。
- 设置分享提取码。
- 查询分享详情。
- 查询分享文件信息。
- 分享文件转存。
- 查询转存任务。
- 获取分享下载地址。
- 获取分享链接音视频文件流数据。
- 外链数据回调。

### 创建分享链接

接口：

```text
POST /apaas/1.0/share/set?product=netdisk
Host: pan.baidu.com
```

参数：

- `appid`
- `access_token`
- `product=netdisk`
- `fsid_list`：文件 ID 字符串数组的 JSON。
- `period`：有效期，单位天。
- `pwd`：4 位提取码，数字加小写字母。
- `remark`
- `ticket`：可选，企业权益。

返回：

- `short_url`
- `link`
- `share_id`
- `period`
- `pwd`
- `remark`

### 分享文件转存

接口：

```text
POST /apaas/1.0/share/transfer?product=netdisk
Host: pan.baidu.com
```

参数：

- `appid`
- `access_token`
- `short_url`
- `fsid_list`
- `to_path`
- `spwd`
- `async`，建议 `2`。
- `ondup`：`fail` 或 `newcopy`。

限制：

- 普通用户单次最多转存 500 个文件。
- 普通会员单次最多转存 3000 个文件。
- 超级会员单次最多转存 50000 个文件。

## 推荐本项目封装结构

后续落代码时建议放在独立包，例如：

```text
internal/infra/baidupan/
  00_client.go
  10_auth.go
  20_user.go
  20_file_list.go
  20_file_meta.go
  20_file_manager.go
  30_upload.go
  30_download.go
  30_share.go
  types.go
```

配置建议：

```go
type BaiduPanConfig struct {
	AppID                int64
	AppKey               string
	SecretKey            string
	SignKey              string
	UserAgent            string
	RequestIntervalMS    int
	FailureRetryMS       int
	MaxRetries           int
	UploadPartSizeBytes  int64
}
```

业务层建议只依赖封装后的接口：

```go
type PanClient interface {
	UserInfo(ctx context.Context, token string) (*UserInfo, error)
	Quota(ctx context.Context, token string) (*Quota, error)
	ListFiles(ctx context.Context, token string, req ListFilesRequest) (*ListFilesResponse, error)
	FileMetas(ctx context.Context, token string, fsids []uint64, opt FileMetaOption) (*FileMetasResponse, error)
	SearchFiles(ctx context.Context, token string, req SearchFilesRequest) (*SearchFilesResponse, error)
	CreateFolder(ctx context.Context, token string, path string, rtype int) (*CreateFileResponse, error)
	MoveFile(ctx context.Context, token string, srcPath, destDir, newName string, ondup string) (*FileManagerResponse, error)
	RenameFile(ctx context.Context, token string, path, newName string, ondup string) (*FileManagerResponse, error)
	DeleteFiles(ctx context.Context, token string, paths []string) (*FileManagerResponse, error)
	UploadFile(ctx context.Context, token string, localPath, remotePath string, opt UploadOption) (*CreateFileResponse, error)
	DownloadFile(ctx context.Context, token string, fsid uint64, dst io.Writer, opt DownloadOption) error
}
```

## 实现自检清单

- 所有远程请求都带 `context.Context`。
- 所有需要鉴权的接口都带 `access_token`。
- 下载请求必须设置 `User-Agent: pan.baidu.com`。
- 上传路径符合 `/apps/{appname}`。
- 上传 `path`、`size`、`isdir` 在预上传和创建文件阶段保持一致。
- `block_list` 在预上传阶段使用本地分片 MD5。
- `block_list` 在创建文件阶段使用分片上传返回的云端 MD5。
- 分片上传使用获取上传域名接口返回的 `https` 域名。
- 文件移动、复制、重命名、删除统一走 `filemanager`。
- 创建文件夹使用 `create` 且 `isdir=1`。
- 每两个远程请求之间按配置 sleep；失败重试前也按配置 sleep。
- 不在业务代码中硬编码请求间隔、重试次数、分片大小之外的隐式业务限制。

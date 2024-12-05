package lark

import (
    "context"
    "github.com/Xhofe/go-cache"
    larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
    log "github.com/sirupsen/logrus"
    "time"
)

const objTokenCacheDuration = 5 * time.Minute
const emptyFolderToken = "empty"

var objTokenCache = cache.NewMemCache[string]()
var exOpts = cache.WithEx[string](objTokenCacheDuration)

func (c *Lark) getObjToken(ctx context.Context, folderID string) (string, bool) {
    if token, ok := objTokenCache.Get(folderID); ok {
        return token, true
    }

    req := larkdrive.NewListFileReqBuilder().FolderToken(folderID).Build()
    resp, err := c.client.Drive.File.ListByIterator(ctx, req)

    if err != nil {
        log.WithError(err).Error("failed to list files")
        return emptyFolderToken, false
    }

    var file *larkdrive.File
    for {
        found, file, err = resp.Next()
        if !found {
            break
        }

        if err != nil {
            log.WithError(err).Error("failed to get next file")
            break
        }

        if *file.Token == folderID {
            objTokenCache.Set(folderID, *file.Token, exOpts)
            return *file.Token, true
        }
    }

    return emptyFolderToken, false
}
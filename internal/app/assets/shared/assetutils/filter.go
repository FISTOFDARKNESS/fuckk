package assetutils

import (
	"github.com/FISTOFDARKNESS/Asset-Reuploader/internal/app/context"
	"github.com/FISTOFDARKNESS/Asset-Reuploader/internal/app/request"
	"github.com/FISTOFDARKNESS/Asset-Reuploader/internal/roblox/develop"
)

func NewFilter(ctx *context.Context, r *request.Request, assetTypeID int32) func(assetsInfo develop.GetAssetsInfoResponse) []*develop.AssetInfo {
	creatorID := r.CreatorID
	userID := ctx.Client.UserInfo.ID
	checkUserID := !r.IsGroup

	return func(assetsInfo develop.GetAssetsInfoResponse) []*develop.AssetInfo {
    filteredAssetsInfo := assetsInfo.Data[:0]
    for _, info := range assetsInfo.Data {
        if info.TypeID != assetTypeID {
            continue
        }

        assetCreatorID := info.Creator.TargetID
        if assetCreatorID == 1 {
            continue // skip Roblox-owned assets
        }

        if assetCreatorID != creatorID && !(checkUserID && assetCreatorID == userID) {
            continue // skip assets not owned by the place creator or logged-in user
        }

        filteredAssetsInfo = append(filteredAssetsInfo, info)
    }
    return filteredAssetsInfo
}
}

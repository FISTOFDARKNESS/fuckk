local Reupload = {}

local AssetIdFilter = require(script.Parent.Parent.App.AssetIdFilter)
local Connection = require(script.Parent.Parent.App.Connection)
local Retry = require(script.Parent.Parent.App.Retry)
local StatusCodes = require(script.Parent.Parent.App.StatusCodes)

local MarketplaceService = game:GetService("MarketplaceService")
local ScriptEditorService = game:GetService("ScriptEditorService")
local HttpService = game:GetService("HttpService")

local _busy = false

type RequestPayload = {
	placeId: number,
	creatorId: number,
	ids: {number},
	defaultPlaceIds: {number},
	pluginVersion: string,
	assetType: string,
	exportJson: boolean,
	isGroup: boolean,
}

local function checkScriptPermission(): boolean
	local testScript = Instance.new("Script")
	local ok = pcall(ScriptEditorService.UpdateSourceAsync, ScriptEditorService, testScript, function() return "" end)
	testScript:Destroy()
	return ok
end

local function getGameInfo(): (boolean, number, number)
	local creatorId = game.CreatorId
	local placeId = game.PlaceId
	
	if placeId == 0 then
		return false, creatorId, placeId
	end
	
	if creatorId ~= 0 then
		local isGroup = game.CreatorType == Enum.CreatorType.Group
		return isGroup, creatorId, placeId
	end
	
	local ok, info = Retry.attempt(3, MarketplaceService.GetProductInfo, MarketplaceService, placeId)
	if not ok then
		return false, creatorId, placeId
	end
	
	local creator = (info :: any).Creator
	local isGroup = creator.CreatorType == "Group"
	return isGroup, creator.CreatorTargetId, placeId
end

export type ReuploadConfig = {
	plugin: Plugin,
	ui: any,
	filterOpts: AssetIdFilter.FilterOptions,
	assetType: string,
	placeIds: {number}?,
}

function Reupload.execute(config: ReuploadConfig)
	local plugin, ui, filterOpts, assetType, placeIds = 
		config.plugin, config.ui, config.filterOpts, config.assetType, config.placeIds
	
	if _busy then
		ui:Notify("Busy", "Plugin is currently busy.")
		return
	end
	_busy = true
	
	if not checkScriptPermission() then
		ui:Notify("Error", "Script injection permission required.")
		_busy = false
		return
	end
	
	local isGroup, creatorId, placeId = getGameInfo()
	if placeId == 0 then
		ui:Notify("Error", "Game must be published first.")
		_busy = false
		return
	elseif creatorId == 0 then
		ui:Notify("Error", "Failed to get place info.")
		_busy = false
		return
	end
	
	local filtered = AssetIdFilter.filterInstances(filterOpts)
	if not next(filtered) then
		ui:notify("Error", "No asset IDs found.")
		_busy = false
		return
	end
	
	local port = plugin:GetSetting("Port") or 3000
	local conn, initialData = Connection.create(port)
	if not conn then
		ui:Notify("Error", "Please initialize the .exe")
		_busy = false
		return
	end
	
	local function onDataReceived(data: string)
		local ok, decoded = pcall(HttpService.JSONDecode, HttpService, data)
		if not ok then
			if data == "done" then
				ui:Notify("Complete", "Finished reuploading all IDs.")
				_busy = false
				Connection.destroy(conn)
			end
			return
		end
		
		if #decoded > 0 then
			AssetIdFilter.replaceIds(filtered, decoded)
			warn(`Replaced {#decoded} IDs`)
		end
	end
	
	conn.onData = onDataReceived
	conn.onDisconnect = function()
		ui:Notify("Disconnected", "Lost connection to localhost.")
		_busy = false
	end
	
	local version = script.Parent.Parent.Parent.Version.Value
	local payload: RequestPayload = {
		pluginVersion = version,
		creatorId = creatorId,
		isGroup = isGroup,
		assetType = assetType,
		ids = AssetIdFilter.getIdArray(filtered),
		placeId = placeId,
		defaultPlaceIds = placeIds or {},
		exportJson = plugin:GetSetting("ExportJSON") or false,
	}
	
	local ok, statusCode = Connection.send(conn, "/reupload", HttpService:JSONEncode(payload))
	if not ok then
		local msg = StatusCodes[tonumber(statusCode) or 0] or `Connection failed: {statusCode}`
		ui:Notify("Error", msg)
		_busy = false
		Connection.destroy(conn)
		return
	end
	
	ui:Notify("Sent", `Sent {#payload.ids} IDs to localhost.`)
	
	if initialData then
		onDataReceived(initialData)
	end
end

return Reupload
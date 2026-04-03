--!strict
local AssetIdFilter = require(script.Parent.Parent.modules.AssetIdFilter)
local Reupload = require(script.Parent.Parent.modules.Reupload)
local ApiDump = require(script.Parent.Parent.modules.ApiDump)
local Selection = game:GetService("Selection")

local function setupDecalTab(ui, tab, plugin)
    local split1 = tab:CreateSplit()
    split1:CreateToggle("Decals", function(state)
        plugin:SetSetting("Decals", state)
    end, plugin:GetSetting("Decals"))
    split1:CreateToggle("Textures", function(state)
        plugin:SetSetting("Textures", state)
    end, plugin:GetSetting("Textures"))
    
    local split2 = tab:CreateSplit()
    split2:CreateToggle("GUI Decals", function(state)
        plugin:SetSetting("GUIDecals", state)
    end, plugin:GetSetting("GUIDecals"))
    split2:CreateToggle("Surface Decals", function(state)
        plugin:SetSetting("SurfaceDecals", state)
    end, plugin:GetSetting("SurfaceDecals"))
    
    tab:CreateButton("Reupload All", function()
        local filter = {
            WhitelistedInstances = {},
            Instances = game:GetDescendants(),
        }
        
        local decalSettings = {
            Decals = "Decal",
            Textures = "Texture",
        }
        
        for setting, className in decalSettings do
            if plugin:GetSetting(setting) then
                table.insert(filter.WhitelistedInstances, className)
            end
        end
        
        local settings = {StringValue = "StrVals", NumberValue = "NumVals", IntValue = "NumVals"}
        for className, setting in settings do
            if plugin:GetSetting(setting) then
                table.insert(filter.whitelist, className)
            end
        end
        if plugin:GetSetting("Scripts") then
            table.insert(filter.whitelist, "LuaSourceContainer")
        end
        
        Reupload.execute({
            plugin = plugin,
            ui = ui,
            filterOpts = filter,
            assetType = "Decal",
            placeIds = _G.getPlaceList and _G.getPlaceList() or {},
        })
    end)
    
    tab:CreateButton("Reupload Selected", function()
        local filter = {
            WhitelistedInstances = {"Decal", "Texture", "StringValue", "NumberValue", "IntValue", "LuaSourceContainer"},
            Instances = Selection:Get(),
        }
        
        Reupload.execute({
            plugin = plugin,
            ui = ui,
            filterOpts = filter,
            assetType = "Decal",
            placeIds = _G.getPlaceList and _G.getPlaceList() or {},
        })
    end)
end

return function(ui, plugin)
    local tab = ui:CreateTab("Decal")
    
    if not ApiDump.isCached() then
        tab:CreateButton("Loading API dump...", function() end)
        task.spawn(function()
            ApiDump.get()
            for _, child in tab.panel:GetChildren() do
                if child:IsA("GuiObject") then child:Destroy() end
            end
            tab._elements = {}
            setupDecalTab(ui, tab, plugin)
        end)
    else
        setupDecalTab(ui, tab, plugin)
    end
end

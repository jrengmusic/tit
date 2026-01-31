package ui

// SpringTheme is a spring-themed color palette with greens and vibrant energy
const SpringTheme = `name = "Spring"
description = "Fresh spring greens with vibrant energy"

[palette]
# Backgrounds - sapphire → ceruleanBlue → sapphire gradient
mainBackgroundColor = "#323B9E"       # sapphire (main background)
inlineBackgroundColor = "#0972BB"     # easternBlue (secondary areas)
selectionBackgroundColor = "#090D12"  # bunker (highlight areas)

# Text - Content & Body - green colors for positive, red for negative
contentTextColor = "#179CA8"          # easternBlue - neutral readable
labelTextColor = "#90D88D"            # feijoa (labels)
dimmedTextColor = "#C8E189"           # yellowGreen (dimmed)
accentTextColor = "#FEEA85"           # salomie - bright shortcuts
highlightTextColor = "#D1D5DA"        # off-white (highlights)
terminalTextColor = "#999999"         # neutral gray

# Special Text
cwdTextColor = "#FEEA85"              # salomie - bright yellow accent
footerTextColor = "#58C9BA"           # downy - muted descriptions

# Borders
boxBorderColor = "#90D88D"            # feijoa
separatorColor = "#0972BB"            # easternBlue

# Confirmation Dialog
confirmationDialogBackground = "#244DA8"  # ceruleanBlue

# Conflict Resolver - Borders
conflictPaneUnfocusedBorder = "#2C4144"
conflictPaneFocusedBorder = "#90D88D"

# Conflict Resolver - Selection
conflictSelectionForeground = "#323B9E"  # sapphire
conflictSelectionBackground = "#7EB8C5"  # brighter muted teal

# Conflict Resolver - Pane Headers
conflictPaneTitleColor = "#90D88D"       # feijoa

# Status Colors - green for clean, red for dirty
statusClean = "#5BCF90"               # emerald - vibrant positive
statusDirty = "#FD5B68"               # wildWatermelon (dirty = red)

# Timeline Colors
timelineSynchronized = "#4ECB71"      # emerald (synced)
timelineLocalAhead = "#5BCF90"        # emerald (ahead)
timelineLocalBehind = "#F67F78"       # froly (behind)

# Operation Colors - green for positive operations
operationReady = "#4ECB71"            # emerald (ready)
operationNotRepo = "#FD5B68"          # wildWatermelon (not repo)
operationTimeTravel = "#F19A84"       # apricot (time travel)
operationConflicted = "#FD5B68"       # wildWatermelon (conflicted)
operationMerging = "#5BCF90"          # emerald (merging)
operationRebasing = "#5BCF90"         # emerald (rebasing)
operationDirtyOp = "#FD5B68"          # wildWatermelon (dirty)

# UI Elements
menuSelectionBackground = "#5BCF90"   # emerald - natural green
buttonSelectedTextColor = "#3F2894"   # daisyBush - dark contrast

# Animation
spinnerColor = "#00FFFF"              # electric cyan

# Diff Colors
diffAddedLineColor = "#5BCF90"        # emerald (added)
diffRemovedLineColor = "#FD5B68"      # wildWatermelon (removed)

# Console Output Colors
outputStdoutColor = "#999999"         # neutral gray
outputStderrColor = "#FD5B68"         # wildWatermelon
outputStatusColor = "#4ECB71"         # emerald
outputWarningColor = "#F67F78"        # froly
outputDebugColor = "#C8E189"          # yellowGreen
outputInfoColor = "#37CB9F"           # shamrock
`

// SummerTheme is a summer-themed color palette with electric blues and bright sunshine
const SummerTheme = `name = "Summer"
description = "Warm summer blues and bright sunshine"

[palette]
# Backgrounds - blueMarguerite → havelockBlue → violetBlue
mainBackgroundColor = "#000000"       # black (main background)
inlineBackgroundColor = "#4D88D1"     # havelockBlue (secondary areas)
selectionBackgroundColor = "#090D12"  # bunker (highlight areas)

# Text - Content & Body - electric cyan/bright for positives, hot reds for negatives
contentTextColor = "#3CA7E0"          # violetBlue - readable neutral
labelTextColor = "#19E5FF"            # cyan (labels)
dimmedTextColor = "#5E68C1"           # indigo (dimmed)
accentTextColor = "#FFBF16"           # lightningYellow - electric shortcuts
highlightTextColor = "#D1D5DA"        # off-white (highlights)
terminalTextColor = "#999999"         # neutral gray

# Special Text
cwdTextColor = "#FFBF16"              # lightningYellow - electric accent
footerTextColor = "#8667BF"           # blueMarguerite - muted descriptions

# Borders
boxBorderColor = "#19E5FF"            # cyan
separatorColor = "#4D88D1"            # havelockBlue

# Confirmation Dialog
confirmationDialogBackground = "#2BC6F0"  # pictonBlue

# Conflict Resolver - Borders
conflictPaneUnfocusedBorder = "#2C4144"
conflictPaneFocusedBorder = "#19E5FF"

# Conflict Resolver - Selection
conflictSelectionForeground = "#3CA7E0"   # pictonBlue
conflictSelectionBackground = "#7EB8C5"   # brighter muted teal

# Conflict Resolver - Pane Headers
conflictPaneTitleColor = "#19E5FF"        # cyan

# Status Colors - electric cyan for clean, hot red for dirty
statusClean = "#19E5FF"               # cyan - electric positive
statusDirty = "#FF3469"               # radicalRed (dirty)

# Timeline Colors
timelineSynchronized = "#00FFFF"      # electric cyan (synced)
timelineLocalAhead = "#19E5FF"        # cyan (ahead)
timelineLocalBehind = "#FF9700"       # pizazz (behind)

# Operation Colors - electric colors for positive ops
operationReady = "#00FFFF"            # electric cyan (ready)
operationNotRepo = "#FF3469"          # radicalRed (not repo)
operationTimeTravel = "#FFBF16"       # lightningYellow (time travel)
operationConflicted = "#FF3469"       # radicalRed (conflicted)
operationMerging = "#19E5FF"          # cyan (merging)
operationRebasing = "#19E5FF"         # cyan (rebasing)
operationDirtyOp = "#FF3469"          # radicalRed (dirty)

# UI Elements
menuSelectionBackground = "#FE62B9"   # hotPink - electric highlight
buttonSelectedTextColor = "#8667BF"   # blueMarguerite - dark contrast

# Animation
spinnerColor = "#00FFFF"              # electric cyan

# Diff Colors
diffAddedLineColor = "#19E5FF"        # cyan (added)
diffRemovedLineColor = "#FF3469"      # radicalRed (removed)

# Console Output Colors
outputStdoutColor = "#999999"         # neutral gray
outputStderrColor = "#FF3469"         # radicalRed
outputStatusColor = "#00FFFF"         # electric cyan
outputWarningColor = "#FF9700"        # pizazz
outputDebugColor = "#5E68C1"          # indigo
outputInfoColor = "#2BC6F0"           # pictonBlue
`

// AutumnTheme is an autumn-themed color palette with rich golds and warm earth tones
const AutumnTheme = `name = "Autumn"
description = "Rich autumn oranges and warm earth tones"

[palette]
# Backgrounds - jacaranda → mulberryWood → roseBudCherry
mainBackgroundColor = "#3E0338"       # jacaranda (main background)
inlineBackgroundColor = "#5E063E"     # mulberryWood (secondary areas)
selectionBackgroundColor = "#090D12"  # bunker (highlight areas)

# Text - Content & Body - gold colors for positive, deep reds for negative
contentTextColor = "#E78C79"          # apricot - warm readable
labelTextColor = "#F9C94D"            # saffronMango (labels)
dimmedTextColor = "#F09D06"           # tulipTree (dimmed)
accentTextColor = "#F5BB09"           # corn - bright shortcuts
highlightTextColor = "#D1D5DA"        # off-white (highlights)
terminalTextColor = "#999999"         # neutral gray

# Special Text
cwdTextColor = "#F5BB09"              # corn - golden bright
footerTextColor = "#CD5861"           # chestnutRose - muted descriptions

# Borders
boxBorderColor = "#F9C94D"            # saffronMango
separatorColor = "#5E063E"            # mulberryWood

# Confirmation Dialog
confirmationDialogBackground = "#7D0E36"  # roseBudCherry

# Conflict Resolver - Borders
conflictPaneUnfocusedBorder = "#2C4144"
conflictPaneFocusedBorder = "#F9C94D"

# Conflict Resolver - Selection
conflictSelectionForeground = "#3E0338"   # jacaranda
conflictSelectionBackground = "#7EB8C5"   # brighter muted teal

# Conflict Resolver - Pane Headers
conflictPaneTitleColor = "#F9C94D"        # saffronMango

# Status Colors - gold for clean, deep red for dirty
statusClean = "#F5BB09"               # corn (clean = gold)
statusDirty = "#DC3003"               # grenadier (dirty = deep red)

# Timeline Colors
timelineSynchronized = "#F5BB09"      # corn (synced)
timelineLocalAhead = "#F9C94D"        # saffronMango (ahead)
timelineLocalBehind = "#E85C03"       # trinidad (behind)

# Operation Colors - gold colors for positive ops
operationReady = "#F5BB09"            # corn (ready)
operationNotRepo = "#DC3003"          # grenadier (not repo)
operationTimeTravel = "#F2AB53"       # safflower (time travel)
operationConflicted = "#DC3003"       # grenadier (conflicted)
operationMerging = "#F5BB09"          # corn (merging)
operationRebasing = "#F5BB09"         # corn (rebasing)
operationDirtyOp = "#DC3003"          # grenadier (dirty)

# UI Elements
menuSelectionBackground = "#F1AE37"   # tulipTree - golden harvest highlight
buttonSelectedTextColor = "#3E0338"   # jacaranda - darkest contrast

# Animation
spinnerColor = "#00FFFF"              # electric cyan

# Diff Colors
diffAddedLineColor = "#F5BB09"        # corn (added)
diffRemovedLineColor = "#DC3003"      # grenadier (removed)

# Console Output Colors
outputStdoutColor = "#999999"         # neutral gray
outputStderrColor = "#DC3003"         # grenadier
outputStatusColor = "#F5BB09"         # corn
outputWarningColor = "#E85C03"        # trinidad
outputDebugColor = "#F09D06"          # tulipTree
outputInfoColor = "#F48C06"           # tangerine
`

// WinterTheme is a winter-themed color palette with professional blues and subtle elegance
const WinterTheme = `name = "Winter"
description = "Cool winter purples with subtle elegance"

[palette]
# Backgrounds - cloudBurst → sanJuan → sanMarino
mainBackgroundColor = "#233253"       # cloudBurst (main background)
inlineBackgroundColor = "#334676"     # sanJuan (secondary areas)
selectionBackgroundColor = "#090D12"  # bunker (highlight areas)

# Text - Content & Body - professional blues for positive, soft pinks for negative
contentTextColor = "#CAD0E6"          # cyanGray - cool readable
labelTextColor = "#7F95D6"            # chetwodeBlue (labels)
dimmedTextColor = "#9BA9D0"           # rockBlue (dimmed)
accentTextColor = "#F6F5FA"           # whisper - bright shortcuts
highlightTextColor = "#D1D5DA"        # off-white (highlights)
terminalTextColor = "#999999"         # neutral gray

# Special Text
cwdTextColor = "#F6F5FA"              # whisper - bright white
footerTextColor = "#9BA9D0"           # rockBlue - muted descriptions

# Borders
boxBorderColor = "#7F95D6"            # chetwodeBlue
separatorColor = "#334676"            # sanJuan

# Confirmation Dialog
confirmationDialogBackground = "#233253"  # cloudBurst

# Conflict Resolver - Borders
conflictPaneUnfocusedBorder = "#2C4144"
conflictPaneFocusedBorder = "#7F95D6"

# Conflict Resolver - Selection
conflictSelectionForeground = "#233253"   # cloudBurst
conflictSelectionBackground = "#7EB8C5"   # brighter muted teal

# Conflict Resolver - Pane Headers
conflictPaneTitleColor = "#7F95D6"        # chetwodeBlue

# Status Colors - professional blue for clean, soft pink for dirty
statusClean = "#6281DC"               # havelockBlue - professional positive
statusDirty = "#E0BACF"               # melanie (dirty = soft pink)

# Timeline Colors
timelineSynchronized = "#435A98"      # sanMarino (synced)
timelineLocalAhead = "#6281DC"        # havelockBlue (ahead)
timelineLocalBehind = "#CEBAC5"       # lily (behind)

# Operation Colors - professional blue colors for positive ops
operationReady = "#435A98"            # sanMarino (ready)
operationNotRepo = "#E0BACF"          # melanie (not repo)
operationTimeTravel = "#CEBAC5"       # lily (time travel)
operationConflicted = "#E0BACF"       # melanie (conflicted)
operationMerging = "#6281DC"          # havelockBlue (merging)
operationRebasing = "#6281DC"         # havelockBlue (rebasing)
operationDirtyOp = "#E0BACF"          # melanie (dirty)

# UI Elements
menuSelectionBackground = "#7F95D6"   # chetwodeBlue - professional blue accent
buttonSelectedTextColor = "#F6F5FA"   # whisper - light contrast

# Animation
spinnerColor = "#00FFFF"              # electric cyan

# Diff Colors
diffAddedLineColor = "#6281DC"        # havelockBlue (added)
diffRemovedLineColor = "#E0BACF"      # melanie (removed)

# Console Output Colors
outputStdoutColor = "#999999"         # neutral gray
outputStderrColor = "#E0BACF"         # melanie
outputStatusColor = "#435A98"         # sanMarino
outputWarningColor = "#CEBAC5"        # lily
outputDebugColor = "#9BA9D0"          # rockBlue
outputInfoColor = "#435A98"           # sanMarino
`

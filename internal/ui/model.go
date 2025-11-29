package ui

import (
	"strconv"
	"time"

	"github.com/davidbudnick/redis/internal/cmd"
	"github.com/davidbudnick/redis/internal/types"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Screen            types.Screen
	Connections       []types.Connection
	SelectedConnIdx   int
	ConnInputs        []textinput.Model
	ConnFocusIdx      int
	EditingConnection *types.Connection
	CurrentConn       *types.Connection
	Keys              []types.RedisKey
	SelectedKeyIdx    int
	KeyCursor         uint64
	KeyPattern        string
	PatternInput      textinput.Model
	CurrentKey        *types.RedisKey
	CurrentValue      types.RedisValue
	AddKeyInputs      []textinput.Model
	AddKeyFocusIdx    int
	AddKeyType        types.KeyType
	TTLInput          textinput.Model
	ServerInfo        types.ServerInfo
	TotalKeys         int64
	Width             int
	Height            int
	Err               error
	StatusMsg         string
	Loading           bool
	ConfirmType       string
	ConfirmData       interface{}
	Logs              *[]string

	// New fields for additional features
	EditValueInput     textinput.Model
	EditingIndex       int
	EditingField       string
	AddCollectionInput []textinput.Model
	AddCollFocusIdx    int
	RenameInput        textinput.Model
	CopyInput          textinput.Model
	SearchValueInput   textinput.Model
	ExportInput        textinput.Model
	ImportInput        textinput.Model
	LuaScriptInput     textinput.Model
	LuaResult          string
	PubSubInput        []textinput.Model
	PubSubFocusIdx     int
	PubSubMessages     []types.PubSubMessage
	SlowLogEntries     []types.SlowLogEntry
	MemoryUsage        int64
	SelectedItemIdx    int
	SortBy             string
	SortAsc            bool
	DBSwitchInput      textinput.Model
	TestConnResult     string
	LogCursor          int
	ShowingLogDetail   bool

	// Favorites and recent keys
	Favorites         []types.Favorite
	RecentKeys        []types.RecentKey
	SelectedFavIdx    int
	SelectedRecentIdx int

	// Tree view
	TreeNodes       []types.TreeNode
	TreeExpanded    map[string]bool
	TreeSeparator   string
	SelectedTreeIdx int

	// Bulk operations
	BulkDeleteInput   textinput.Model
	BulkDeletePreview []string
	BulkDeleteCount   int
	SelectedBulkKeys  map[string]bool

	// Batch TTL
	BatchTTLInput   textinput.Model
	BatchTTLPattern textinput.Model
	BatchTTLPreview []string

	// Search
	RegexSearchInput textinput.Model
	FuzzySearchInput textinput.Model
	SearchResults    []types.RedisKey

	// Watch mode
	WatchActive     bool
	WatchKey        string
	WatchValue      string
	WatchLastUpdate time.Time
	WatchInterval   time.Duration

	// Client list and memory stats
	ClientList        []types.ClientInfo
	MemoryStats       *types.MemoryStats
	SelectedClientIdx int

	// Cluster mode
	ClusterNodes    []types.ClusterNode
	ClusterEnabled  bool
	SelectedNodeIdx int

	// Compare keys
	CompareKey1Input textinput.Model
	CompareKey2Input textinput.Model
	CompareResult    *types.KeyComparison
	CompareFocusIdx  int

	// Key templates
	Templates           []types.KeyTemplate
	SelectedTemplateIdx int
	TemplateInputs      []textinput.Model
	TemplateFocusIdx    int

	// JSON path query
	JSONPathInput  textinput.Model
	JSONPathResult string

	// Keybindings
	KeyBindings types.KeyBindings

	// Value history
	ValueHistory       []types.ValueHistoryEntry
	SelectedHistoryIdx int

	// Keyspace events
	KeyspaceEvents    []types.KeyspaceEvent
	KeyspaceSubActive bool
	KeyspacePattern   string

	// Connection groups
	ConnectionGroups []types.ConnectionGroup
	SelectedGroupIdx int

	// Expiring keys alerts
	ExpiringKeys    []types.RedisKey
	ExpiryThreshold int64 // seconds
}

func NewModel() Model {
	patternInput := textinput.New()
	patternInput.Placeholder = "Pattern (e.g., user:*)"
	patternInput.Width = 30

	ttlInput := textinput.New()
	ttlInput.Placeholder = "TTL in seconds (0 = no expiry)"
	ttlInput.Width = 30

	editValueInput := textinput.New()
	editValueInput.Placeholder = "New value"
	editValueInput.Width = 50

	renameInput := textinput.New()
	renameInput.Placeholder = "New key name"
	renameInput.Width = 40

	copyInput := textinput.New()
	copyInput.Placeholder = "Destination key name"
	copyInput.Width = 40

	searchValueInput := textinput.New()
	searchValueInput.Placeholder = "Search for value..."
	searchValueInput.Width = 40

	exportInput := textinput.New()
	exportInput.Placeholder = "export.json"
	exportInput.Width = 40
	exportInput.SetValue("export.json")

	importInput := textinput.New()
	importInput.Placeholder = "import.json"
	importInput.Width = 40

	luaScriptInput := textinput.New()
	luaScriptInput.Placeholder = "return redis.call('PING')"
	luaScriptInput.Width = 60

	dbSwitchInput := textinput.New()
	dbSwitchInput.Placeholder = "Database number (0-15)"
	dbSwitchInput.Width = 30

	// New feature inputs
	bulkDeleteInput := textinput.New()
	bulkDeleteInput.Placeholder = "Pattern to delete (e.g., cache:*)"
	bulkDeleteInput.Width = 40

	batchTTLInput := textinput.New()
	batchTTLInput.Placeholder = "TTL in seconds"
	batchTTLInput.Width = 20

	batchTTLPattern := textinput.New()
	batchTTLPattern.Placeholder = "Key pattern (e.g., session:*)"
	batchTTLPattern.Width = 40

	regexSearchInput := textinput.New()
	regexSearchInput.Placeholder = "Regex pattern (e.g., user:\\d+)"
	regexSearchInput.Width = 40

	fuzzySearchInput := textinput.New()
	fuzzySearchInput.Placeholder = "Fuzzy search term..."
	fuzzySearchInput.Width = 40

	compareKey1Input := textinput.New()
	compareKey1Input.Placeholder = "First key to compare"
	compareKey1Input.Width = 40

	compareKey2Input := textinput.New()
	compareKey2Input.Placeholder = "Second key to compare"
	compareKey2Input.Width = 40

	jsonPathInput := textinput.New()
	jsonPathInput.Placeholder = "JSON path (e.g., $.users[0].name)"
	jsonPathInput.Width = 50

	return Model{
		Screen:             types.ScreenConnections,
		Connections:        []types.Connection{},
		ConnInputs:         createConnectionInputs(),
		AddKeyInputs:       createAddKeyInputs(),
		AddKeyType:         types.KeyTypeString,
		PatternInput:       patternInput,
		TTLInput:           ttlInput,
		Keys:               []types.RedisKey{},
		EditValueInput:     editValueInput,
		AddCollectionInput: createAddCollectionInputs(),
		RenameInput:        renameInput,
		CopyInput:          copyInput,
		SearchValueInput:   searchValueInput,
		ExportInput:        exportInput,
		ImportInput:        importInput,
		LuaScriptInput:     luaScriptInput,
		PubSubInput:        createPubSubInputs(),
		DBSwitchInput:      dbSwitchInput,
		SortBy:             "name",
		SortAsc:            true,

		// Favorites and recent
		Favorites:         []types.Favorite{},
		RecentKeys:        []types.RecentKey{},
		SelectedFavIdx:    0,
		SelectedRecentIdx: 0,

		// Tree view
		TreeNodes:       []types.TreeNode{},
		TreeExpanded:    make(map[string]bool),
		TreeSeparator:   ":",
		SelectedTreeIdx: 0,

		// Bulk operations
		BulkDeleteInput:   bulkDeleteInput,
		BulkDeletePreview: []string{},
		SelectedBulkKeys:  make(map[string]bool),

		// Batch TTL
		BatchTTLInput:   batchTTLInput,
		BatchTTLPattern: batchTTLPattern,
		BatchTTLPreview: []string{},

		// Search
		RegexSearchInput: regexSearchInput,
		FuzzySearchInput: fuzzySearchInput,
		SearchResults:    []types.RedisKey{},

		// Watch mode
		WatchActive:   false,
		WatchInterval: time.Second * 2,

		// Client list and memory
		ClientList:        []types.ClientInfo{},
		SelectedClientIdx: 0,

		// Cluster
		ClusterNodes:    []types.ClusterNode{},
		ClusterEnabled:  false,
		SelectedNodeIdx: 0,

		// Compare keys
		CompareKey1Input: compareKey1Input,
		CompareKey2Input: compareKey2Input,
		CompareFocusIdx:  0,

		// Templates
		Templates:           []types.KeyTemplate{},
		SelectedTemplateIdx: 0,
		TemplateInputs:      []textinput.Model{},
		TemplateFocusIdx:    0,

		// JSON path
		JSONPathInput: jsonPathInput,

		// Keybindings
		KeyBindings: types.DefaultKeyBindings(),

		// Value history
		ValueHistory:       []types.ValueHistoryEntry{},
		SelectedHistoryIdx: 0,

		// Keyspace events
		KeyspaceEvents:    []types.KeyspaceEvent{},
		KeyspaceSubActive: false,

		// Connection groups
		ConnectionGroups: []types.ConnectionGroup{},
		SelectedGroupIdx: 0,

		// Expiring keys
		ExpiringKeys:    []types.RedisKey{},
		ExpiryThreshold: 3600, // 1 hour default
	}
}

func createConnectionInputs() []textinput.Model {
	inputs := make([]textinput.Model, 5)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Connection Name"
	inputs[0].Focus()
	inputs[0].Width = 30

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Host"
	inputs[1].Width = 30
	inputs[1].SetValue("localhost")

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Port"
	inputs[2].Width = 30
	inputs[2].SetValue("6379")

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Password (optional)"
	inputs[3].Width = 30
	inputs[3].EchoMode = textinput.EchoPassword

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Database (0-15)"
	inputs[4].Width = 30
	inputs[4].SetValue("0")

	return inputs
}

func createAddKeyInputs() []textinput.Model {
	inputs := make([]textinput.Model, 2)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Key Name"
	inputs[0].Focus()
	inputs[0].Width = 30

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Value"
	inputs[1].Width = 30

	return inputs
}

func (m Model) Init() tea.Cmd {
	return cmd.LoadConnectionsCmd()
}

func (m Model) getPort() int {
	port, err := strconv.Atoi(m.ConnInputs[2].Value())
	if err != nil {
		return 6379
	}
	return port
}

func (m Model) getDB() int {
	db, err := strconv.Atoi(m.ConnInputs[4].Value())
	if err != nil {
		return 0
	}
	return db
}

func (m *Model) resetConnInputs() {
	for i := range m.ConnInputs {
		m.ConnInputs[i].SetValue("")
		m.ConnInputs[i].Blur()
	}
	m.ConnInputs[1].SetValue("localhost")
	m.ConnInputs[2].SetValue("6379")
	m.ConnInputs[4].SetValue("0")
	m.ConnInputs[0].Focus()
	m.ConnFocusIdx = 0
}

func (m *Model) resetAddKeyInputs() {
	for i := range m.AddKeyInputs {
		m.AddKeyInputs[i].SetValue("")
		m.AddKeyInputs[i].Blur()
	}
	m.AddKeyInputs[0].Focus()
	m.AddKeyFocusIdx = 0
	m.AddKeyType = types.KeyTypeString
}

func (m *Model) populateConnInputs(conn types.Connection) {
	m.ConnInputs[0].SetValue(conn.Name)
	m.ConnInputs[1].SetValue(conn.Host)
	m.ConnInputs[2].SetValue(strconv.Itoa(conn.Port))
	m.ConnInputs[3].SetValue(conn.Password)
	m.ConnInputs[4].SetValue(strconv.Itoa(conn.DB))
}

func createAddCollectionInputs() []textinput.Model {
	inputs := make([]textinput.Model, 2)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Value / Member / Field"
	inputs[0].Focus()
	inputs[0].Width = 40

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Score / Value (optional)"
	inputs[1].Width = 40

	return inputs
}

func createPubSubInputs() []textinput.Model {
	inputs := make([]textinput.Model, 2)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Channel name"
	inputs[0].Focus()
	inputs[0].Width = 40

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Message"
	inputs[1].Width = 40

	return inputs
}

func (m *Model) resetAddCollectionInputs() {
	for i := range m.AddCollectionInput {
		m.AddCollectionInput[i].SetValue("")
		m.AddCollectionInput[i].Blur()
	}
	m.AddCollectionInput[0].Focus()
	m.AddCollFocusIdx = 0
}

func (m *Model) resetPubSubInputs() {
	for i := range m.PubSubInput {
		m.PubSubInput[i].SetValue("")
		m.PubSubInput[i].Blur()
	}
	m.PubSubInput[0].Focus()
	m.PubSubFocusIdx = 0
}

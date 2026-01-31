package graph

import (
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGraph(t *testing.T) {
	g := NewGraph()
	assert.NotNil(t, g)
	assert.NotNil(t, g.nodes)
	assert.NotNil(t, g.edges)
	assert.Empty(t, g.nodes)
	assert.Empty(t, g.edges)
}

func TestAddNode(t *testing.T) {
	t.Run("add single node", func(t *testing.T) {
		g := NewGraph()
		pkg := config.Package{
			Name:      "core",
			Path:      "./core",
			Ecosystem: config.EcosystemGo,
		}

		err := g.AddNode(pkg)
		assert.NoError(t, err)

		// Verify node was added
		node, exists := g.GetNode("core")
		assert.True(t, exists)
		assert.NotNil(t, node)
		assert.Equal(t, "core", node.Package.Name)
	})

	t.Run("add duplicate node returns error", func(t *testing.T) {
		g := NewGraph()
		pkg := config.Package{
			Name:      "core",
			Path:      "./core",
			Ecosystem: config.EcosystemGo,
		}

		// Add first time - should succeed
		err := g.AddNode(pkg)
		require.NoError(t, err)

		// Add second time - should fail
		err = g.AddNode(pkg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestAddEdge(t *testing.T) {
	tests := []struct {
		name     string
		from     string
		to       string
		strategy string
		bumpMap  map[string]string
		wantErr  bool
	}{
		{
			name:     "add edge between existing nodes",
			from:     "api",
			to:       "core",
			strategy: "linked",
			bumpMap:  nil,
			wantErr:  false,
		},
		{
			name:     "add edge from non-existent node",
			from:     "nonexistent",
			to:       "core",
			strategy: "linked",
			bumpMap:  nil,
			wantErr:  true,
		},
		{
			name:     "add edge to non-existent node",
			from:     "api",
			to:       "nonexistent",
			strategy: "linked",
			bumpMap:  nil,
			wantErr:  true,
		},
		{
			name:     "add edge with custom bump mapping",
			from:     "web",
			to:       "api",
			strategy: "linked",
			bumpMap: map[string]string{
				"major": "patch",
				"minor": "patch",
				"patch": "patch",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGraph()
			// Add nodes
			_ = g.AddNode(config.Package{Name: "core"})
			_ = g.AddNode(config.Package{Name: "api"})
			_ = g.AddNode(config.Package{Name: "web"})

			err := g.AddEdge(tt.from, tt.to, tt.strategy, tt.bumpMap)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify edge was added
				edges := g.GetEdgesFrom(tt.from)
				assert.NotEmpty(t, edges)
				found := false
				for _, edge := range edges {
					if edge.To == tt.to {
						found = true
						assert.Equal(t, tt.strategy, edge.Strategy)
						if tt.bumpMap != nil {
							assert.Equal(t, tt.bumpMap, edge.BumpMap)
						}
					}
				}
				assert.True(t, found, "Edge not found in graph")
			}
		})
	}
}

func TestGetNode(t *testing.T) {
	g := NewGraph()
	pkg := config.Package{
		Name:      "core",
		Path:      "./core",
		Ecosystem: config.EcosystemGo,
	}
	err := g.AddNode(pkg)
	require.NoError(t, err)

	// Get existing node
	node, exists := g.GetNode("core")
	assert.True(t, exists)
	assert.NotNil(t, node)
	assert.Equal(t, "core", node.Package.Name)
	assert.Equal(t, 0, node.SCC) // Default SCC value

	// Get non-existent node
	node, exists = g.GetNode("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, node)
}

func TestGetEdgesFrom(t *testing.T) {
	g := NewGraph()

	// Set up graph: api -> core, web -> api
	_ = g.AddNode(config.Package{Name: "core"})
	_ = g.AddNode(config.Package{Name: "api"})
	_ = g.AddNode(config.Package{Name: "web"})

	_ = g.AddEdge("api", "core", "linked", nil)
	_ = g.AddEdge("web", "api", "linked", nil)

	// Get edges from api
	edges := g.GetEdgesFrom("api")
	assert.Len(t, edges, 1)
	assert.Equal(t, "core", edges[0].To)

	// Get edges from web
	edges = g.GetEdgesFrom("web")
	assert.Len(t, edges, 1)
	assert.Equal(t, "api", edges[0].To)

	// Get edges from node with no outgoing edges
	edges = g.GetEdgesFrom("core")
	assert.Empty(t, edges)

	// Get edges from non-existent node
	edges = g.GetEdgesFrom("nonexistent")
	assert.Empty(t, edges)
}

func TestGetAllNodes(t *testing.T) {
	g := NewGraph()

	packages := []config.Package{
		{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
		{Name: "api", Path: "./api", Ecosystem: config.EcosystemNPM},
		{Name: "web", Path: "./web", Ecosystem: config.EcosystemNPM},
	}

	for _, pkg := range packages {
		err := g.AddNode(pkg)
		require.NoError(t, err)
	}

	nodes := g.GetAllNodes()
	assert.Len(t, nodes, 3)

	// Verify all nodes are present
	nodeNames := make(map[string]bool)
	for _, node := range nodes {
		nodeNames[node.Package.Name] = true
	}
	assert.True(t, nodeNames["core"])
	assert.True(t, nodeNames["api"])
	assert.True(t, nodeNames["web"])
}

func TestComplexGraph(t *testing.T) {
	// Test a more complex graph structure
	g := NewGraph()

	// Create packages
	packages := []config.Package{
		{Name: "utils", Path: "./utils", Ecosystem: config.EcosystemGo},
		{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
		{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo},
		{Name: "web", Path: "./web", Ecosystem: config.EcosystemNPM},
		{Name: "mobile", Path: "./mobile", Ecosystem: config.EcosystemNPM},
	}

	for _, pkg := range packages {
		err := g.AddNode(pkg)
		require.NoError(t, err)
	}

	// Create edges: web -> api -> core -> utils
	//               mobile -> api
	edges := []struct {
		from     string
		to       string
		strategy string
	}{
		{"core", "utils", "linked"},
		{"api", "core", "linked"},
		{"web", "api", "linked"},
		{"mobile", "api", "fixed"},
	}

	for _, edge := range edges {
		err := g.AddEdge(edge.from, edge.to, edge.strategy, nil)
		require.NoError(t, err)
	}

	// Verify structure
	assert.Len(t, g.GetAllNodes(), 5)
	assert.Len(t, g.GetEdgesFrom("core"), 1)
	assert.Len(t, g.GetEdgesFrom("api"), 1)
	assert.Len(t, g.GetEdgesFrom("web"), 1)
	assert.Len(t, g.GetEdgesFrom("mobile"), 1)
	assert.Empty(t, g.GetEdgesFrom("utils"))
}

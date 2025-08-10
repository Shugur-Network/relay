package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/Shugur-Network/relay/internal/logger"
	"go.uber.org/zap"
)

// CockroachClusterNode represents a node in the CockroachDB cluster
type CockroachClusterNode struct {
	NodeID        int64     `json:"node_id"`
	Address       string    `json:"address"`
	SQLAddress    string    `json:"sql_address"`
	Locality      string    `json:"locality"`
	ClusterName   string    `json:"cluster_name"`
	ServerVersion string    `json:"server_version"`
	BuildTag      string    `json:"build_tag"`
	StartedAt     time.Time `json:"started_at"`
	IsLive        bool      `json:"is_live"`
	Ranges        int64     `json:"ranges"`
	Leases        int64     `json:"leases"`
}

// CockroachClusterInfo represents cluster summary information
type CockroachClusterInfo struct {
	TotalNodes   int64                   `json:"total_nodes"`
	LiveNodes    int64                   `json:"live_nodes"`
	CurrentNode  *CockroachClusterNode   `json:"current_node"`
	AllNodes     []*CockroachClusterNode `json:"all_nodes"`
	ClusterName  string                  `json:"cluster_name"`
	IsCluster    bool                    `json:"is_cluster"`
}

// GetCockroachClusterInfo retrieves cluster information from CockroachDB
func (db *DB) GetCockroachClusterInfo(ctx context.Context) (*CockroachClusterInfo, error) {
	if !db.isConnected() {
		return nil, fmt.Errorf("database is not connected")
	}

	query := `
		SELECT 
			node_id, 
			address, 
			sql_address, 
			COALESCE(locality, '') as locality,
			COALESCE(cluster_name, '') as cluster_name,
			server_version,
			build_tag,
			started_at, 
			is_live,
			ranges,
			leases
		FROM crdb_internal.gossip_nodes
		ORDER BY node_id
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query cluster nodes: %w", err)
	}
	defer rows.Close()

	var nodes []*CockroachClusterNode
	var liveCount int64
	var clusterName string

	for rows.Next() {
		node := &CockroachClusterNode{}
		err := rows.Scan(
			&node.NodeID,
			&node.Address,
			&node.SQLAddress,
			&node.Locality,
			&node.ClusterName,
			&node.ServerVersion,
			&node.BuildTag,
			&node.StartedAt,
			&node.IsLive,
			&node.Ranges,
			&node.Leases,
		)
		if err != nil {
			logger.Error("Failed to scan cluster node", zap.Error(err))
			continue
		}

		nodes = append(nodes, node)
		if node.IsLive {
			liveCount++
		}
		if clusterName == "" && node.ClusterName != "" {
			clusterName = node.ClusterName
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cluster nodes: %w", err)
	}

	// Get current node information
	currentNode, err := db.getCurrentNode(ctx, nodes)
	if err != nil {
		logger.Warn("Failed to identify current node", zap.Error(err))
	}

	clusterInfo := &CockroachClusterInfo{
		TotalNodes:  int64(len(nodes)),
		LiveNodes:   liveCount,
		CurrentNode: currentNode,
		AllNodes:    nodes,
		ClusterName: clusterName,
		IsCluster:   len(nodes) > 1, // More than 1 node means it's a cluster
	}

	return clusterInfo, nil
}

// getCurrentNode attempts to identify which node we're currently connected to
func (db *DB) getCurrentNode(ctx context.Context, nodes []*CockroachClusterNode) (*CockroachClusterNode, error) {
	// Try to get the current node ID from CockroachDB
	query := `SELECT crdb_internal.node_id()`
	
	var currentNodeID int64
	err := db.Pool.QueryRow(ctx, query).Scan(&currentNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current node ID: %w", err)
	}

	// Find the current node in the list
	for _, node := range nodes {
		if node.NodeID == currentNodeID {
			return node, nil
		}
	}

	return nil, fmt.Errorf("current node ID %d not found in cluster nodes", currentNodeID)
}

// GetClusterHealth returns cluster health information
func (db *DB) GetClusterHealth(ctx context.Context) (map[string]interface{}, error) {
	if !db.isConnected() {
		return nil, fmt.Errorf("database is not connected")
	}

	health := make(map[string]interface{})

	// Get cluster summary
	clusterInfo, err := db.GetCockroachClusterInfo(ctx)
	if err != nil {
		return nil, err
	}

	health["total_nodes"] = clusterInfo.TotalNodes
	health["live_nodes"] = clusterInfo.LiveNodes
	health["dead_nodes"] = clusterInfo.TotalNodes - clusterInfo.LiveNodes
	health["cluster_name"] = clusterInfo.ClusterName
	health["is_cluster"] = clusterInfo.IsCluster

	if clusterInfo.CurrentNode != nil {
		health["current_node_id"] = clusterInfo.CurrentNode.NodeID
		health["current_node_address"] = clusterInfo.CurrentNode.Address
		health["current_node_live"] = clusterInfo.CurrentNode.IsLive
	}

	// Calculate cluster health percentage
	if clusterInfo.TotalNodes > 0 {
		healthPercent := float64(clusterInfo.LiveNodes) / float64(clusterInfo.TotalNodes) * 100
		health["health_percentage"] = healthPercent
		
		if healthPercent == 100 {
			health["status"] = "healthy"
		} else if healthPercent >= 50 {
			health["status"] = "degraded"
		} else {
			health["status"] = "critical"
		}
	} else {
		health["status"] = "unknown"
		health["health_percentage"] = 0
	}

	return health, nil
}

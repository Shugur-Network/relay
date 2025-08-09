// Shugur Relay Dashboard JavaScript

class RelayDashboard {
  constructor() {
    this.statsUpdateInterval = null;
    this.init();
  }

  init() {
    this.setupEventListeners();
    this.startStatsUpdates();
  }

  // Setup event listeners
  setupEventListeners() {
    // Add smooth scrolling for anchor links
    document.querySelectorAll('a[href^="#"]').forEach((anchor) => {
      anchor.addEventListener("click", function (e) {
        e.preventDefault();
        const target = document.querySelector(this.getAttribute("href"));
        if (target) {
          target.scrollIntoView({
            behavior: "smooth",
            block: "start",
          });
        }
      });
    });

    // Handle visibility change to pause/resume updates when tab is not visible
    document.addEventListener("visibilitychange", () => {
      if (document.hidden) {
        this.stopStatsUpdates();
      } else {
        this.startStatsUpdates();
      }
    });
  }

  // Start automatic stats updates
  startStatsUpdates() {
    // Update immediately
    this.updateStats();
    
    // Add some visual feedback to stats cards
    this.addStatsCardAnimations();
    
    // Update every 5 seconds for real-time feel
    this.statsUpdateInterval = setInterval(() => {
      this.updateStats();
    }, 5000);
  }

  // Add hover animations and visual feedback to stats cards
  addStatsCardAnimations() {
    const statCards = document.querySelectorAll('.stat-card');
    
    statCards.forEach(card => {
      // Add subtle pulse animation to show they're "live"
      card.style.transition = 'transform 0.2s ease, box-shadow 0.2s ease';
      
      card.addEventListener('mouseenter', () => {
        card.style.transform = 'translateY(-2px)';
        card.style.boxShadow = '0 8px 25px rgba(0, 0, 0, 0.15)';
      });
      
      card.addEventListener('mouseleave', () => {
        card.style.transform = 'translateY(0)';
        card.style.boxShadow = '';
      });
      
      // Add a subtle indicator that shows the card is updating
      const statValue = card.querySelector('.stat-value');
      if (statValue) {
        setInterval(() => {
          statValue.style.textShadow = '0 0 10px rgba(59, 130, 246, 0.3)';
          setTimeout(() => {
            statValue.style.textShadow = '';
          }, 500);
        }, 30000); // Pulse every 30 seconds
      }
    });
  }

  // Stop automatic stats updates
  stopStatsUpdates() {
    if (this.statsUpdateInterval) {
      clearInterval(this.statsUpdateInterval);
      this.statsUpdateInterval = null;
    }
  }

  // Update statistics by fetching from API
  async updateStats() {
    try {
      const response = await fetch('/api/stats');
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      
      const data = await response.json();
      
      // Update all the stats with real data
      if (data.stats) {
        this.updateStatElement('active-connections', data.stats.active_connections);
        this.updateStatElement('messages-processed', data.stats.messages_processed);
        this.updateStatElement('events-stored', data.stats.events_stored);
      }
      
      // Update uptime
      if (data.uptime) {
        this.updateStatElement('uptime', data.uptime);
      }
      
      // Update online indicator
      this.updateOnlineIndicator(true);
      this.addLastUpdatedIndicator();
      
    } catch (error) {
      console.warn('Failed to update stats:', error);
      // Update the status indicator to show offline
      this.updateOnlineIndicator(false);
    }
  }

  // Update individual stat element with animation
  updateStatElement(elementId, newValue) {
    const element = document.getElementById(elementId);
    if (!element) return;

    const currentValue = element.textContent.trim();
    let newValueStr;
    
    // Format the value based on type
    if (elementId === 'uptime') {
      newValueStr = newValue; // Uptime is already formatted
    } else {
      newValueStr = this.formatStatValue(newValue);
    }
    
    if (currentValue !== newValueStr) {
      // Add a subtle animation when value changes
      element.style.transform = 'scale(1.05)';
      element.style.transition = 'transform 0.2s ease-in-out';
      element.style.color = '#3b82f6'; // Brief blue highlight
      
      setTimeout(() => {
        element.textContent = newValueStr;
        element.style.transform = 'scale(1)';
        element.style.color = ''; // Reset color
      }, 100);
      
      // Remove transition after animation
      setTimeout(() => {
        element.style.transition = '';
      }, 300);
      
      // Add a subtle pulse to show it updated
      setTimeout(() => {
        element.style.textShadow = '0 0 8px rgba(59, 130, 246, 0.4)';
        setTimeout(() => {
          element.style.textShadow = '';
        }, 1000);
      }, 200);
    }
  }

  // Format stat values for display
  formatStatValue(value) {
    // If it's a number, format with commas
    if (typeof value === 'number') {
      return value.toLocaleString();
    }
    return value;
  }

  // Update online indicator
  updateOnlineIndicator(isOnline) {
    const statusDot = document.querySelector('.status-dot');
    const statusText = document.querySelector('.status-indicator span');
    
    if (statusDot && statusText) {
      if (isOnline) {
        statusDot.className = 'status-dot online';
        statusText.textContent = 'Online';
      } else {
        statusDot.className = 'status-dot offline';
        statusText.textContent = 'Offline';
      }
    }
  }

  // Add a "last updated" indicator to show the dashboard is live
  addLastUpdatedIndicator() {
    const now = new Date();
    const timeString = now.toLocaleTimeString();
    
    // Find or create the last updated indicator
    let indicator = document.getElementById('last-updated');
    if (!indicator) {
      indicator = document.createElement('div');
      indicator.id = 'last-updated';
      indicator.style.cssText = `
        position: fixed;
        bottom: 20px;
        right: 20px;
        background: rgba(0, 0, 0, 0.7);
        color: white;
        padding: 8px 12px;
        border-radius: 8px;
        font-size: 12px;
        font-family: monospace;
        z-index: 1000;
        opacity: 0;
        transition: opacity 0.3s ease;
      `;
      document.body.appendChild(indicator);
    }
    
    indicator.textContent = `Last updated: ${timeString}`;
    indicator.style.opacity = '1';
    
    // Fade out after 2 seconds
    setTimeout(() => {
      indicator.style.opacity = '0.3';
    }, 2000);
  }
}

// Utility Functions

// Copy text to clipboard
function copyToClipboard(elementId) {
  const element = document.getElementById(elementId);
  if (!element) return;

  const text = element.textContent;

  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard
      .writeText(text)
      .then(() => {
        showToast("Copied to clipboard!");
      })
      .catch((err) => {
        console.error("Failed to copy text: ", err);
        fallbackCopy(text);
      });
  } else {
    fallbackCopy(text);
  }
}

// Fallback copy method for older browsers
function fallbackCopy(text) {
  const textArea = document.createElement("textarea");
  textArea.value = text;
  textArea.style.position = "fixed";
  textArea.style.top = "0";
  textArea.style.left = "0";
  textArea.style.width = "2em";
  textArea.style.height = "2em";
  textArea.style.padding = "0";
  textArea.style.border = "none";
  textArea.style.outline = "none";
  textArea.style.boxShadow = "none";
  textArea.style.background = "transparent";

  document.body.appendChild(textArea);
  textArea.focus();
  textArea.select();

  try {
    document.execCommand("copy");
    showToast("Copied to clipboard!");
  } catch (err) {
    console.error("Fallback copy failed: ", err);
    showToast("Failed to copy to clipboard");
  }

  document.body.removeChild(textArea);
}

// Show toast notification
function showToast(message, type = "success", duration = 3000) {
  // Remove existing toast
  const existingToast = document.querySelector(".toast");
  if (existingToast) {
    existingToast.remove();
  }

  // Create toast element
  const toast = document.createElement("div");
  toast.className = `toast toast-${type}`;
  toast.textContent = message;

  // Add toast styles
  Object.assign(toast.style, {
    position: "fixed",
    top: "20px",
    right: "20px",
    padding: "12px 24px",
    borderRadius: "12px",
    color: "white",
    fontWeight: "500",
    zIndex: "10000",
    transform: "translateX(100%)",
    transition: "transform 0.3s ease-in-out",
    backgroundColor: type === "success" ? "#22c55e" : "#ef4444",
    boxShadow: "0 10px 30px rgba(0, 0, 0, 0.3)",
  });

  document.body.appendChild(toast);

  // Animate in
  setTimeout(() => {
    toast.style.transform = "translateX(0)";
  }, 100);

  // Remove after duration
  setTimeout(() => {
    toast.style.transform = "translateX(100%)";
    setTimeout(() => {
      if (toast.parentNode) {
        toast.parentNode.removeChild(toast);
      }
    }, 300);
  }, duration);
}

// Format uptime
function formatUptime(seconds) {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  if (days > 0) {
    return `${days}d ${hours}h ${minutes}m`;
  } else if (hours > 0) {
    return `${hours}h ${minutes}m`;
  } else {
    return `${minutes}m`;
  }
}

// Initialize dashboard when DOM is loaded
document.addEventListener("DOMContentLoaded", () => {
  new RelayDashboard();
  new CockroachClusterInfo();

  // Set WebSocket URL dynamically
  const websocketUrlElement = document.getElementById("websocket-url");
  if (websocketUrlElement) {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    websocketUrlElement.textContent = `${protocol}//${window.location.host}`;
  }
});

// CockroachDB Cluster Information Handler
class CockroachClusterInfo {
  constructor() {
    this.tooltip = null;
    this.init();
  }

  init() {
    this.createTooltip();
    this.attachEventListeners();
  }

  createTooltip() {
    this.tooltip = document.createElement('div');
    this.tooltip.className = 'cluster-tooltip';
    this.tooltip.style.cssText = `
      position: absolute;
      background: #fafafa;
      color: #e2e8f0;
      padding: 12px;
      border-radius: 8px;
      box-shadow: 0 4px 20px rgba(0,0,0,0.4);
      z-index: 1000;
      display: none;
      max-width: 400px;
      font-size: 13px;
      line-height: 1.4;
      border: 1px solid #4a5568;
    `;
    document.body.appendChild(this.tooltip);
  }

  attachEventListeners() {
    document.addEventListener('mouseenter', (e) => {
      if (e.target.closest('.cluster-indicator')) {
        this.showClusterTooltip(e);
      }
    }, true);

    document.addEventListener('mouseleave', (e) => {
      if (e.target.closest('.cluster-indicator')) {
        this.hideClusterTooltip();
      }
    }, true);
  }

  async showClusterTooltip(event) {
    try {
      const response = await fetch(`/api/cluster`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      
      const clusterInfo = await response.json();
      
      this.tooltip.innerHTML = this.renderClusterTooltipContent(clusterInfo);
      this.positionTooltip(event);
      this.tooltip.style.display = 'block';
    } catch (error) {
      console.error('Error fetching cluster information:', error);
      this.tooltip.innerHTML = `
        <div class="cluster-tooltip-header">
          <strong>Cluster Information</strong>
        </div>
        <div class="cluster-error">
          Failed to load cluster information
        </div>
      `;
      this.positionTooltip(event);
      this.tooltip.style.display = 'block';
    }
  }

  renderClusterTooltipContent(clusterInfo) {
    let content = `
      <div class="cluster-summary">
        <div class="cluster-stat">
          <span class="stat-label">Live Nodes:</span>
          <span class="stat-value live-nodes">${clusterInfo.live_nodes}/${clusterInfo.total_nodes}</span>
        </div>
      </div>
      <div class="cluster-nodes-list">
    `;

    if (clusterInfo.all_nodes && clusterInfo.all_nodes.length > 0) {
      clusterInfo.all_nodes.forEach(node => {
        const isCurrentNode = clusterInfo.current_node && node.node_id === clusterInfo.current_node.node_id;
        const statusClass = node.is_live ? 'live' : 'dead';
        
        // Extract hostname from address (remove port)
        let hostname = node.address;
        if (hostname && hostname.includes(':')) {
          hostname = hostname.split(':')[0];
        }
        
        content += `
          <div class="cluster-node-minimal ${statusClass}">
            <div class="node-minimal-info">
              <span class="node-hostname">${hostname || 'Unknown'}</span>
              <span class="node-status-minimal ${statusClass}">${node.is_live ? 'LIVE' : 'DOWN'}</span>
            </div>
            <div class="node-started-minimal">Started: ${this.formatTimestamp(node.started_at)}</div>
          </div>
        `;
      });
    }

    content += '</div>';
    return content;
  }

  positionTooltip(event) {
    const rect = event.target.getBoundingClientRect();
    const tooltipRect = this.tooltip.getBoundingClientRect();
    
    let left = rect.left + window.scrollX;
    let top = rect.bottom + window.scrollY + 8;
    
    // Adjust if tooltip goes off screen
    if (left + tooltipRect.width > window.innerWidth) {
      left = window.innerWidth - tooltipRect.width - 10;
    }
    
    if (top + tooltipRect.height > window.innerHeight + window.scrollY) {
      top = rect.top + window.scrollY - tooltipRect.height - 8;
    }
    
    this.tooltip.style.left = `${left}px`;
    this.tooltip.style.top = `${top}px`;
  }

  hideClusterTooltip() {
    this.tooltip.style.display = 'none';
  }

  formatTimestamp(timestamp) {
    try {
      const date = new Date(timestamp);
      return date.toLocaleString();
    } catch (error) {
      return 'Unknown';
    }
  }
}

// Add CSS for toast notifications
const toastStyle = document.createElement("style");
toastStyle.textContent = `
    .toast {
        font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    }
`;
document.head.appendChild(toastStyle);

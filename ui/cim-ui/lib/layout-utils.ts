import { Node, Edge } from "./mock-data";

/**
 * A simple Force-Directed Layout algorithm.
 * Arranges nodes in a circle initially and then applies 
 * repulsive and attractive forces for a specified number of iterations.
 */
export const arrangeNodes = (nodes: Node[], edges: Edge[]) => {
  const iterations = 50;
  const k = Math.sqrt((800 * 600) / nodes.length); // Optimal distance
  
  // Clone nodes to avoid mutating original data
  const layoutNodes = nodes.map((node, i) => ({
    ...node,
    // If no coordinates, place in a circle
    x: node.x ?? 400 + 200 * Math.cos((2 * Math.PI * i) / nodes.length),
    y: node.y ?? 300 + 200 * Math.sin((2 * Math.PI * i) / nodes.length),
    vx: 0,
    vy: 0,
  }));

  for (let iter = 0; iter < iterations; iter++) {
    // 1. Repulsive forces between all nodes
    for (let i = 0; i < layoutNodes.length; i++) {
      for (let j = 0; j < layoutNodes.length; j++) {
        if (i === j) continue;
        const dx = layoutNodes[i].x - layoutNodes[j].x;
        const dy = layoutNodes[i].y - layoutNodes[j].y;
        const distance = Math.sqrt(dx * dx + dy * dy) || 1;
        const force = (k * k) / distance;
        layoutNodes[i].vx += (dx / distance) * force;
        layoutNodes[i].vy += (dy / distance) * force;
      }
    }

    // 2. Attractive forces along edges
    edges.forEach((edge) => {
      const source = layoutNodes.find((n) => n.id === edge.from);
      const target = layoutNodes.find((n) => n.id === edge.to);
      if (source && target) {
        const dx = target.x - source.x;
        const dy = target.y - source.y;
        const distance = Math.sqrt(dx * dx + dy * dy) || 1;
        const force = (distance * distance) / k;
        const fx = (dx / distance) * force;
        const fy = (dy / distance) * force;
        source.vx += fx;
        source.vy += fy;
        target.vx -= fx;
        target.vy -= fy;
      }
    });

    // 3. Apply displacement
    layoutNodes.forEach((node) => {
      const damping = 0.1;
      node.x += node.vx * damping;
      node.y += node.vy * damping;
      // Reset velocity
      node.vx = 0;
      node.vy = 0;
    });
  }

  // Final Pass: Snap to a grid to help orthogonal routing look cleaner
  return layoutNodes.map(node => ({
    ...node,
    x: Math.round(node.x / 50) * 50,
    y: Math.round(node.y / 50) * 50
  }));
};

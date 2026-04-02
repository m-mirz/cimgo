export interface NodeAttribute {
  [key: string]: string | number | boolean;
}

export interface Node {
  id: string;
  x?: number;
  y?: number;
  attributes: NodeAttribute;
}

export interface Edge {
  from: string;
  to: string;
  attributes?: NodeAttribute;
}

export interface GraphData {
  nodes: Node[];
  edges: Edge[];
}

export interface ModelInfo {
  id: string;
  name: string;
  isCustom?: boolean;
}

export const INITIAL_MODELS: ModelInfo[] = [
  { id: "grid1", name: "Grid 1" },
  { id: "grid2", name: "Grid 2" },
  { id: "dynamic", name: "Dynamic Layout (No Coords)" },
];

// In-memory storage for models and custom data
// In a real app, this would be a database
export let MODELS: ModelInfo[] = [...INITIAL_MODELS];
export const CUSTOM_DATA: Record<string, GraphData> = {};

export const addCustomModel = (name: string, data: GraphData) => {
  const id = `custom-${Date.now()}`;
  const newModel = { id, name, isCustom: true };
  MODELS = [...MODELS, newModel];
  CUSTOM_DATA[id] = data;
  return newModel;
};

export const generateMockData = (modelId: string = "production"): GraphData => {
  // Return custom data if it exists
  if (CUSTOM_DATA[modelId]) {
    return CUSTOM_DATA[modelId];
  }

  const nodes: Node[] = [];
  const edges: Edge[] = [];
  
  const counts: Record<string, number> = {
    grid1: 30,
    grid2: 15,
    dynamic: 25,
  };
  
  const nodeCount = counts[modelId] || 20;
  const cols = Math.ceil(Math.sqrt(nodeCount * 1.5));
  
  let seed = modelId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
  const pseudoRandom = () => {
    seed = (seed * 9301 + 49297) % 233280;
    return seed / 233280;
  };
  
  const types = ["PowerNode", "Bus", "Info"];
  const statuses = ["Healthy", "Warning", "Critical"];

  for (let i = 0; i < nodeCount; i++) {
    const row = Math.floor(i / cols);
    const col = i % cols;
    
    const x = 100 + col * 200 + pseudoRandom() * 50;
    const y = 100 + row * 150 + pseudoRandom() * 50;
    
    let type = types[Math.floor(pseudoRandom() * (types.length - 2))];
    if (i > 0 && i % 7 === 0) type = "Bus";
    else if (i > 0 && i % 5 === 0) type = "Info";
    
    const status = statuses[Math.floor(pseudoRandom() * statuses.length)];

    nodes.push({
      id: `node-${i + 1}`,
      ...(modelId !== "dynamic" ? { x, y } : {}),
      attributes: {
        "Name": `${type}-${i + 1}`,
        "Type": type,
        "Status": status,
        "Capacity": `${Math.floor(pseudoRandom() * 100)}%`,
      },
    });
  }

  for (let i = 0; i < nodeCount - 1; i++) {
    const createEdge = (from: string, to: string) => {
      const isLineType = pseudoRandom() > 0.7;
      const type = isLineType ? "PowerLine" : "LogicalLink";
      const status = statuses[Math.floor(pseudoRandom() * statuses.length)];
      
      edges.push({
        from,
        to,
        attributes: {
          "Name": `${type}`,
          "Type": type,
          "Status": status,
          "Capacity": `${Math.floor(pseudoRandom() * 100)}%`,
        }
      });
    };

    if (pseudoRandom() > 0.4) {
      createEdge(`node-${i + 1}`, `node-${i + 2}`);
    }
    
    if (i + cols < nodeCount && pseudoRandom() > 0.6) {
      createEdge(`node-${i + 1}`, `node-${i + cols + 1}`);
    }
  }

  return { nodes, edges };
};

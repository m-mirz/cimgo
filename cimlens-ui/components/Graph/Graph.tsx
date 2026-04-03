"use client";

import { useEffect, useState, useRef } from "react";
import { GraphData, Node, Edge, ModelInfo } from "@/lib/mock-data";
import { arrangeNodes } from "@/lib/layout-utils";
import { GraphNode } from "./Node";
import { GraphEdge } from "./Edge";
import { Loader2, ZoomIn, ZoomOut, RotateCcw, Upload } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { AnimatePresence, motion, useMotionValue } from "framer-motion";

export const Graph = () => {
  const [data, setData] = useState<GraphData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [mounted, setMounted] = useState(false);
  const [models, setModels] = useState<ModelInfo[]>([]);
  const [selectedModel, setSelectedModel] = useState("");
  
  // Unified hover state
  const [hoveredItem, setHoveredItem] = useState<{
    type: 'node' | 'edge';
    data: Node | Edge;
    x: number;
    y: number;
  } | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const svgRef = useRef<SVGSVGElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const x = useMotionValue(0);
  const y = useMotionValue(0);
  const scale = useMotionValue(1);

  // Initial models fetch
  useEffect(() => {
    setMounted(true);
    const fetchModels = async () => {
      try {
        const response = await fetch("/api/models");
        const json = await response.json();
        setModels(json);
        if (json.length > 0) setSelectedModel(json[0].id);
      } catch (err) {
        console.error("Failed to fetch models", err);
      }
    };
    fetchModels();
  }, []);

  // Data fetch when model changes
  useEffect(() => {
    if (!selectedModel) return;
    
    const fetchData = async () => {
      setLoading(true);
      try {
        const response = await fetch(`/api/graph?model=${selectedModel}`);
        if (!response.ok) throw new Error("Failed to fetch graph data");
        const json = await response.json() as GraphData;
        
        const needsLayout = json.nodes.some(n => n.x === undefined || n.y === undefined);
        const processedNodes = needsLayout 
          ? arrangeNodes(json.nodes, json.edges)
          : json.nodes;

        setData({ ...json, nodes: processedNodes });
        x.set(0);
        y.set(0);
        scale.set(1);
      } catch (err) {
        setError(err instanceof Error ? err.message : "An unknown error occurred");
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [selectedModel]);

  const handleFileUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = async (e) => {
      try {
        const json = JSON.parse(e.target?.result as string);
        const response = await fetch("/api/models", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            name: file.name.replace(".json", ""),
            data: json
          })
        });

        if (!response.ok) throw new Error("Upload failed");
        
        const newModel = await response.json();
        
        // Refresh models and select the new one
        const modelsRes = await fetch("/api/models");
        const updatedModels = await modelsRes.json();
        setModels(updatedModels);
        setSelectedModel(newModel.id);
      } catch (err) {
        alert("Error uploading model: " + (err instanceof Error ? err.message : "Invalid JSON"));
      }
    };
    reader.readAsText(file);
  };

  const handleZoom = (delta: number, clientX?: number, clientY?: number) => {
    const currentScale = scale.get();
    const newScale = Math.min(Math.max(currentScale + delta, 0.2), 3);
    
    if (newScale === currentScale) return;

    if (clientX !== undefined && clientY !== undefined && containerRef.current) {
      const rect = containerRef.current.getBoundingClientRect();
      
      // Mouse position relative to the container
      const mouseX = clientX - rect.left;
      const mouseY = clientY - rect.top;

      // Current content position relative to the mouse
      const currentX = x.get();
      const currentY = y.get();

      // Calculate how much we need to shift x and y to keep the mouse point stable
      // Formula: new_offset = mouse_pos - (mouse_pos - old_offset) * (new_scale / old_scale)
      const ratio = newScale / currentScale;
      const newX = mouseX - (mouseX - currentX) * ratio;
      const newY = mouseY - (mouseY - currentY) * ratio;

      x.set(newX);
      y.set(newY);
    }
    
    scale.set(newScale);
  };

  const resetTransform = () => {
    if (!containerRef.current || !data || data.nodes.length === 0) {
      x.set(0);
      y.set(0);
      scale.set(1);
      return;
    }

    const containerRect = containerRef.current.getBoundingClientRect();
    const padding = 40; // Padding around the diagram
    const availableWidth = containerRect.width - padding * 2;
    const availableHeight = containerRect.height - padding * 2;

    const nodes = data.nodes;
    const minX = Math.min(...nodes.map(n => n.x ?? 0)) - 100;
    const minY = Math.min(...nodes.map(n => n.y ?? 0)) - 100;
    const maxX = Math.max(...nodes.map(n => n.x ?? 0)) + 100;
    const maxY = Math.max(...nodes.map(n => n.y ?? 0)) + 100;
    const contentWidth = maxX - minX;
    const contentHeight = maxY - minY;

    // Calculate scale to fit
    const scaleX = availableWidth / contentWidth;
    const scaleY = availableHeight / contentHeight;
    const newScale = Math.min(scaleX, scaleY, 1); // Don't zoom in past 100% on reset

    // Calculate offsets to center
    // Centered position = (container_dim / 2) - (content_dim * scale / 2) - (min_coord * scale)
    const newX = (containerRect.width / 2) - (contentWidth * newScale / 2) - (minX * newScale);
    const newY = (containerRect.height / 2) - (contentHeight * newScale / 2) - (minY * newScale);

    scale.set(newScale);
    x.set(newX);
    y.set(newY);
  };

  const getScreenPosition = (itemX: number, itemY: number) => {
    if (!svgRef.current || !containerRef.current || !data) return { left: 0, top: 0 };
    
    // Use the same bounds as the SVG viewbox logic
    const nodes = data.nodes;
    const minX = Math.min(...nodes.map(n => n.x ?? 0)) - 100;
    const minY = Math.min(...nodes.map(n => n.y ?? 0)) - 100;
    const maxX = Math.max(...nodes.map(n => n.x ?? 0)) + 100;
    const maxY = Math.max(...nodes.map(n => n.y ?? 0)) + 100;
    const width = maxX - minX;
    const height = maxY - minY;

    const viewBox = svgRef.current.viewBox.baseVal;
    
    // Scale factor from SVG coordinates to CSS pixels (at scale 1)
    // We need to account for the actual rendered size of the SVG
    const svgRect = svgRef.current.getBoundingClientRect();
    const currentScale = scale.get();
    
    // The scale-independent size of the SVG element
    const baseSvgWidth = svgRect.width / currentScale;
    const baseSvgHeight = svgRect.height / currentScale;

    const internalScaleX = baseSvgWidth / viewBox.width;
    const internalScaleY = baseSvgHeight / viewBox.height;
    
    const currentX = x.get();
    const currentY = y.get();

    const localX = (itemX - viewBox.x) * internalScaleX;
    const localY = (itemY - viewBox.y) * internalScaleY;

    return {
      left: localX * currentScale + currentX + 16,
      top: localY * currentScale + currentY + 16,
    };
  };

  const currentPos = hoveredItem ? getScreenPosition(hoveredItem.x, hoveredItem.y) : null;

  const handleWheel = (e: React.WheelEvent) => {
    e.preventDefault();
    const delta = e.deltaY > 0 ? -0.15 : 0.15;
    handleZoom(delta, e.clientX, e.clientY);
  };

  if (error) {
    return (
      <div className="flex h-full w-full items-center justify-center text-red-500">
        Error: {error}
      </div>
    );
  }

  const nodes = data?.nodes || [];
  const minX = data ? Math.min(...nodes.map(n => n.x ?? 0)) - 100 : 0;
  const minY = data ? Math.min(...nodes.map(n => n.y ?? 0)) - 100 : 0;
  const maxX = data ? Math.max(...nodes.map(n => n.x ?? 0)) + 100 : 0;
  const maxY = data ? Math.max(...nodes.map(n => n.y ?? 0)) + 100 : 0;
  const width = maxX - minX;
  const height = maxY - minY;

  return (
    <div className="relative flex flex-col h-full w-full bg-white dark:bg-zinc-950 overflow-hidden">
      <div className="p-4 border-b border-zinc-100 dark:border-zinc-900 bg-zinc-50/50 dark:bg-zinc-900/50 flex flex-wrap gap-4 justify-between items-center shrink-0">
        <div className="flex items-center gap-4">
          <div>
            <h2 className="text-sm font-semibold text-zinc-700 dark:text-zinc-300">Model</h2>
          </div>
          
          <div className="flex items-center gap-2">
            <Select 
              value={selectedModel} 
              onValueChange={(val) => val && setSelectedModel(val)}
            >
              <SelectTrigger className="w-[180px] h-9 text-xs">
                <SelectValue placeholder="Select Environment" />
              </SelectTrigger>
              {mounted && (
                <SelectContent>
                  {models.map((model) => (
                    <SelectItem key={model.id} value={model.id} className="text-xs">
                      {model.name} {model.isCustom && "(Custom)"}
                    </SelectItem>
                  ))}
                </SelectContent>
              )}
            </Select>

            <input
              type="file"
              ref={fileInputRef}
              onChange={handleFileUpload}
              accept=".json"
              className="hidden"
            />
            <Button 
              variant="outline" 
              size="sm" 
              className="h-9 gap-2 text-xs"
              onClick={() => fileInputRef.current?.click()}
            >
              <Upload className="h-3.5 w-3.5" />
              Upload
            </Button>
          </div>
        </div>

        <div className="flex gap-1 shrink-0">
          <Button variant="outline" size="icon-sm" onClick={() => handleZoom(0.2)} title="Zoom In">
            <ZoomIn className="h-4 w-4" />
          </Button>
          <Button variant="outline" size="icon-sm" onClick={() => handleZoom(-0.2)} title="Zoom Out">
            <ZoomOut className="h-4 w-4" />
          </Button>
          <Button variant="outline" size="icon-sm" onClick={resetTransform} title="Reset View">
            <RotateCcw className="h-4 w-4" />
          </Button>
        </div>
      </div>
      
      <div 
        ref={containerRef}
        onWheel={handleWheel}
        className="relative flex-1 w-full bg-zinc-50/30 dark:bg-zinc-900/10 cursor-grab active:cursor-grabbing overflow-hidden"
      >
        {loading && (
          <div className="absolute inset-0 z-10 flex items-center justify-center bg-white/50 backdrop-blur-[2px] dark:bg-zinc-950/50">
            <Loader2 className="h-8 w-8 animate-spin text-zinc-400" />
          </div>
        )}

        {data && (
          <motion.div
            style={{ x, y, scale, transformOrigin: "0 0" }}
            drag
            dragConstraints={{ left: -width, right: width, top: -height, bottom: height }}
            className="inline-block p-4"
          >
            <svg
              ref={svgRef}
              viewBox={`${minX} ${minY} ${width} ${height}`}
              width={width}
              height={height}
              xmlns="http://www.w3.org/2000/svg"
              className="pointer-events-auto"
              >
              {data.edges.map((edge, idx) => {
                const fromNode = data.nodes.find(n => n.id === edge.from);
                const toNode = data.nodes.find(n => n.id === edge.to);
                if (!fromNode || !toNode || fromNode.x === undefined || toNode.x === undefined) return null;
                return (
                  <GraphEdge 
                    key={`edge-${idx}`} 
                    fromNode={fromNode as any} 
                    toNode={toNode as any} 
                    edge={edge}
                    onHover={(e, ex, ey) => {
                      if (e) setHoveredItem({ type: 'edge', data: e, x: ex, y: ey });
                      else setHoveredItem(null);
                    }}
                  />
                );
              })}

              {data.nodes.map((node) => (
                node.x !== undefined && node.y !== undefined && (
                  <GraphNode 
                    key={node.id} 
                    node={node as any} 
                    onHover={(n) => {
                      if (n) setHoveredItem({ type: 'node', data: n, x: n.x!, y: n.y! });
                      else setHoveredItem(null);
                    }} 
                  />
                )
              ))}
            </svg>
          </motion.div>
        )}

        <AnimatePresence>
          {hoveredItem && currentPos && (
            <motion.div
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.95 }}
              transition={{ duration: 0.1 }}
              className="absolute pointer-events-none z-50"
              style={{
                left: `${currentPos.left}px`,
                top: `${currentPos.top - 10}px`,
                transform: 'translate(-50%, -100%)',
              }}
            >
              <Card className="w-64 shadow-2xl border-zinc-200 bg-white/95 backdrop-blur-sm dark:bg-zinc-950/95 dark:border-zinc-800 overflow-hidden">
                <CardHeader className="pb-2 p-4 bg-zinc-50/50 dark:bg-zinc-900/50 border-b border-zinc-100 dark:border-zinc-800">
                  <div className="flex justify-between items-start gap-2">
                    <CardTitle className="text-sm font-bold truncate">
                      {hoveredItem.data.attributes!["Name"] as string}
                    </CardTitle>
                    <Badge variant="outline" className="text-[10px] px-1.5 py-0 h-5 shrink-0">
                      {hoveredItem.data.attributes!["Type"] as string}
                    </Badge>
                  </div>
                </CardHeader>
                <CardContent className="space-y-2 p-4 pt-3 text-[12px]">
                  <div className="flex justify-between items-center py-1">
                    <span className="text-zinc-500">Status:</span>
                    <span className={`font-semibold ${
                      hoveredItem.data.attributes!["Status"] === "Critical" ? "text-red-500" : 
                      hoveredItem.data.attributes!["Status"] === "Warning" ? "text-yellow-600 dark:text-yellow-400" : 
                      "text-green-600 dark:text-green-400"
                    }`}>{hoveredItem.data.attributes!["Status"] as string}</span>
                  </div>
                  {Object.entries(hoveredItem.data.attributes!).map(([key, value]) => {
                    if (["Name", "Type", "Status"].includes(key)) return null;
                    return (
                      <div key={key} className="flex justify-between items-center py-0.5">
                        <span className="text-zinc-500">{key}:</span>
                        <span className="font-mono text-zinc-900 dark:text-zinc-100">{String(value)}</span>
                      </div>
                    );
                  })}
                </CardContent>
              </Card>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </div>
  );
};

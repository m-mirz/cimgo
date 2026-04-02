"use client";

import { generateOrthogonalPath } from "@/lib/graph-utils";
import { Node, Edge } from "@/lib/mock-data";
import { motion } from "framer-motion";

interface GraphEdgeProps {
  fromNode: Node;
  toNode: Node;
  edge: Edge;
  onHover: (edge: Edge | null, x: number, y: number) => void;
}

export const GraphEdge = ({ fromNode, toNode, edge, onHover }: GraphEdgeProps) => {
  const pathData = generateOrthogonalPath(
    fromNode.x ?? 0,
    fromNode.y ?? 0,
    toNode.x ?? 0,
    toNode.y ?? 0
  );

  const isLineType = edge.attributes?.["Type"] === "PowerLine";

  const handlePointerEnter = (e: React.PointerEvent) => {
    // For edges, we use the midpoint of the path for the tooltip
    const midX = ((fromNode.x ?? 0) + (toNode.x ?? 0)) / 2;
    const midY = ((fromNode.y ?? 0) + (toNode.y ?? 0)) / 2;
    onHover(edge, midX, midY);
  };

  return (
    <g
      onPointerEnter={handlePointerEnter}
      onPointerLeave={() => onHover(null, 0, 0)}
      className="cursor-pointer"
    >
      {/* Invisible thicker path for easier hovering */}
      <path
        d={pathData}
        fill="none"
        stroke="transparent"
        strokeWidth={15}
        className="pointer-events-auto"
      />
      
      <motion.path
        initial={{ pathLength: 0, opacity: 0 }}
        animate={{ pathLength: 1, opacity: 1 }}
        transition={{ duration: 1, ease: "easeInOut" }}
        d={pathData}
        fill="none"
        stroke="currentColor"
        strokeWidth={isLineType ? 2 : 1}
        className={`${
          isLineType 
            ? "text-zinc-500 dark:text-zinc-300" 
            : "text-zinc-300 dark:text-zinc-600"
        } transition-colors duration-200 hover:text-blue-400`}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </g>
  );
};

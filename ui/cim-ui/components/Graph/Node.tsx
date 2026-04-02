"use client";

import { Node } from "@/lib/mock-data";
import { motion } from "framer-motion";

interface GraphNodeProps {
  node: Node;
  onHover: (node: Node | null) => void;
}

export const GraphNode = ({ node, onHover }: GraphNodeProps) => {
  const { attributes, x = 0, y = 0 } = node;
  const status = attributes["Status"] as string;
  const type = attributes["Type"] as string;
  const isBus = type === "Bus";
  const isInfo = type === "Info";

  const getStatusStroke = (status: string) => {
    switch (status) {
      case "Critical": return "stroke-red-500";
      case "Warning": return "stroke-yellow-500";
      default: return "stroke-zinc-300";
    }
  };

  return (
    <motion.g
      initial={{ scale: 0 }}
      animate={{ scale: 1 }}
      whileHover={{ scale: isBus ? 1.05 : (isInfo ? 1.4 : 1.2) }}
      className="cursor-pointer"
      onPointerEnter={() => onHover(node)}
      onPointerLeave={() => onHover(null)}
      style={{ pointerEvents: 'all' }}
    >
      {isBus ? (
        // Bus: Thick horizontal line
        <rect
          x={x - 60}
          y={y - 4}
          width={120}
          height={8}
          rx={4}
          className="fill-zinc-900 dark:fill-zinc-100"
        />
      ) : isInfo ? (
        // Info: Lighter, subtle dot without circle
        <circle
          cx={x}
          cy={y}
          r={8}
          className="fill-zinc-300 dark:fill-zinc-600"
        />
      ) : (
        // Standard Node: Nested circles
        <>
          <circle
            cx={x}
            cy={y}
            r={12}
            className={`fill-white stroke-1 ${getStatusStroke(status)} dark:fill-zinc-900`}
          />
          <circle
            cx={x}
            cy={y}
            r={8}
            className={`fill-zinc-900 dark:fill-zinc-100`}
          />
        </>
      )}
    </motion.g>
  );
};


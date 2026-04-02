/**
 * Generates an SVG path string for an orthogonal connection between two points.
 * Uses a horizontal -> vertical -> horizontal sequence (Z-shape).
 */
export const generateOrthogonalPath = (
  startX: number,
  startY: number,
  endX: number,
  endY: number
): string => {
  // Use a midpoint for the first horizontal turn
  const midX = startX + (endX - startX) / 2;
  
  // M startX startY - Move to start
  // H midX           - Horizontal line to midX
  // V endY           - Vertical line to endY
  // H endX           - Horizontal line to endX
  return `M ${startX} ${startY} H ${midX} V ${endY} H ${endX}`;
};

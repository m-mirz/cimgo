import { NextRequest, NextResponse } from "next/server";
import { generateMockData } from "@/lib/mock-data";

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const modelId = searchParams.get("model") || "production";
  
  const data = generateMockData(modelId);
  return NextResponse.json(data);
}

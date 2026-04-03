import { NextRequest, NextResponse } from "next/server";
import { MODELS, addCustomModel } from "@/lib/mock-data";

export async function GET() {
  return NextResponse.json(MODELS);
}

export async function POST(request: NextRequest) {
  try {
    const { name, data } = await request.json();
    
    if (!name || !data || !data.nodes) {
      return NextResponse.json({ error: "Invalid model data" }, { status: 400 });
    }

    const newModel = addCustomModel(name, data);
    return NextResponse.json(newModel);
  } catch (error) {
    return NextResponse.json({ error: "Failed to process upload" }, { status: 500 });
  }
}

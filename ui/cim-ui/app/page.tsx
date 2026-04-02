import { Graph } from "@/components/Graph/Graph";
import { ModeToggle } from "@/components/ModeToggle";
import { Search } from "lucide-react";

export default function Home() {
  return (
    <div className="flex flex-col h-screen bg-zinc-50 dark:bg-black font-sans overflow-hidden">
      <header className="p-6 border-b border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-950 shrink-0 flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100 flex items-center gap-2">
            CIMlens
            <Search className="h-6 w-6 text-zinc-400" strokeWidth={2.5} />
          </h1>
        </div>
        <ModeToggle />
      </header>
      
      <main className="flex-1 min-h-0">
        <Graph />
      </main>
    </div>
  );
}

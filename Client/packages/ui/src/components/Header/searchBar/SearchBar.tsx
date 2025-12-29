import { Input } from "@workspace/ui/components/input"

import { FaMagnifyingGlass } from 'react-icons/fa6';
function SearchBar() {
  return (
    <div className="relative w-full max-w-md group">
      <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
        <FaMagnifyingGlass className="h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors" />
      </div>
      <Input
        className="
          w-full pl-10 pr-4 h-10
          bg-muted/50 hover:bg-muted/80 focus:bg-background
          border-transparent focus:border-primary/20
          rounded-full
          transition-all duration-200
          placeholder:text-muted-foreground/70
          focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:ring-offset-0
        "
        placeholder="Search markets, events..."
      />
      <div className="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none opacity-0 group-focus-within:opacity-100 transition-opacity">
        <kbd className="pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground opacity-100">
          <span className="text-xs">⌘</span>K
        </kbd>
      </div>
    </div>

  )
}

export default SearchBar

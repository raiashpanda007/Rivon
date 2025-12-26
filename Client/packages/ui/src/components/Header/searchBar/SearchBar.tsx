import { Input } from "@workspace/ui/components/input"
import { Button } from "@workspace/ui/components/button"
import { FaMagnifyingGlass } from 'react-icons/fa6';
function SearchBar() {
  return (
    <div className="w-3/6 h-full flex items-center space-x-1">
      <Input
        className="
          w-5/6
          text-white placeholder:text-gray-400
          border-none outline-none ring-0
          focus-visible:outline-none focus-visible:ring-0
          focus-visible:shadow-[0_0_0_2px_rgba(249,115,22,0.8)]
          transition-shadow duration-200
          bg-[#202127]
        "
        placeholder="Search..."
      />
      <Button className="cursor-pointer">
        <FaMagnifyingGlass className="text-orange-500" />
      </Button>
    </div>

  )
}

export default SearchBar

import { Button } from "@workspace/ui/components/button"
import { FaUser } from "react-icons/fa"

interface UserProfileProps {
  url?: string
}

function UserProfile({ url }: UserProfileProps) {
  return (
    <Button
      variant="outline"
      className="h-12 w-12 rounded-full p-0 overflow-hidden flex items-center justify-center"
    >
      {url ? (
        <img
          src={url}
          alt="User profile"
          className="h-full w-full object-cover"
        />

      ) : (
        <FaUser className="text-lg text-white" />
      )}
    </Button>
  )
}

export default UserProfile

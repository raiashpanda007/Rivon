import { Button } from "@workspace/ui/components/button"
import { FaUser } from "react-icons/fa"

interface UserProfileProps {
  url?: string
}


function UserProfile({ url }: UserProfileProps) {
  const safeUrl = url?.trim()

  return (
    <Button
      variant="outline"
      className="h-12 w-12 rounded-full p-0 overflow-hidden flex items-center justify-center"
    >
      {safeUrl ? (
        <img
          src={safeUrl}
          alt="User profile"
          referrerPolicy="no-referrer"
          className="h-full w-full object-cover"
          onError={() => console.error("Failed to load image:", safeUrl)}
        />
      ) : (
        <FaUser className="text-lg" />
      )}
    </Button>
  )
}
export default UserProfile

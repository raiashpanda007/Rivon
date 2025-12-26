
import { Button } from "@workspace/ui/components/button";
import { FaSignOutAlt } from "react-icons/fa";
function SignOut() {
  return (
    <Button variant={"destructive"} className="font-body font-semibold cursor-pointer">
      Sign Out <FaSignOutAlt />
    </Button>
  )

}



export default SignOut;

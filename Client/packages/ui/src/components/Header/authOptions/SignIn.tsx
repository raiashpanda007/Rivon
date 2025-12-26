
import { Button } from "@workspace/ui/components/button";
import { FaUser } from "react-icons/fa";
function SignIn() {
  return (
    <Button variant={"secondary"} className="font-body font-semibold cursor-pointer">
      Sign in <FaUser />
    </Button>
  )

}



export default SignIn;

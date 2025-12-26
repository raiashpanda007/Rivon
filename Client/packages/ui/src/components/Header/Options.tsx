import { BsGraphUp } from 'react-icons/bs';
import { MdAttachMoney } from 'react-icons/md';
import { Button } from "@workspace/ui/components/button";

function Options() {
  return (
    <div className="w-1/6  flex items-center justify-evenly">
      <Button className="cursor-pointer font-bold font-body" variant={"secondary"} >
        <MdAttachMoney /> Betting
      </Button>
      <Button className="cursor-pointer font-bold font-body bg-orange-500 hover:opacity-80 hover:bg-orange-500" variant={"default"} >
        <BsGraphUp className='font-bold' /> Trading
      </Button>
    </div >
  )
}



export default Options;

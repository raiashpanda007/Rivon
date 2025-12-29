import { Wallet } from "lucide-react"

function WalletIcon() {
  return (
    <div className="h-11/12 opacity-85 border  flex items-center space-x-1">
      <Wallet className="text-orange-500" />
      <span className="font-semibold">
        $ 00.00
      </span>
    </div>
  )
}



export default WalletIcon;

import Logo from "@workspace/ui/components/Logo/Logo";
import AuthOptions from "@workspace/ui/components/Header/authOptions/Authoptions";
import Options from "@workspace/ui/components/Header/Options";
import SearchBar from "@workspace/ui/components/Header/searchBar/SearchBar";
function Header() {
  return (
    <div className="w-full h-1/12 border  border-black flex">
      <Logo className="h-full w-1/6" />
      <Options />
      <SearchBar />
      <AuthOptions />
    </div>
  )

}

export default Header;

import express from "express";
import cors from "cors";
import path from "path";
import dotenv from "dotenv";
import MailHandler from "./Controllers/mailHandler";
dotenv.config();
const PORT = process.env.PORT ?? "8001";

class App {
  public app: express.Application;


  constructor() {
    this.app = express();
    this.IntializeMiddleware();
    this.InitializeRouters()
  }

  public IntializeMiddleware() {
    this.app.use(cors({
      origin: "*",
      credentials: true
    }));
    this.app.use(express.json())
    this.app.use(express.urlencoded({ extended: false }))
  }
  public InitializeRouters() {
    this.app.get("/", (req: express.Request, res: express.Response): void => {
      res.sendFile(path.join(__dirname, "Template.html"))
    })
    this.app.post("/", MailHandler);
  }

  listen() {
    this.app.listen(PORT, () => {
      console.info("Mailing server is up and running you send request on port :: ", PORT);
    })
  }



}


const app = new App();

app.listen();


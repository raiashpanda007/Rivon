import { z as zod } from "zod";
import dotenv from "dotenv"

dotenv.config();

export interface ConfigType {
  PORT: number,
  REDIS_URL: string
}

const ENVSchema = zod.object({
  PORT: zod.coerce.number().int().positive(),
  REDIS_URL: zod.string(),
})

class Config {
  public MustLoad(): ConfigType {
    const parsed = ENVSchema.safeParse(process.env);
    if (!parsed.success) {
      console.error("Invalid environment variables:", parsed.error.format());
      process.exit(1);
    }
    return parsed.data;
  }
}

export default Config;

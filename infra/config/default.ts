import * as dotenv from "dotenv";
dotenv.config();

const envOrElse = (name: string, elseVal: string): string => {
  const val = process.env[name];
  return val ?? elseVal;
};

export interface Config {
  imageUrl: string;
}

export const config: Config = {
  imageUrl: envOrElse("IMAGE_URL", ""),
};

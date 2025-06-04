import type { Route } from "./+types/home";
import { Welcome } from "../welcome/welcome";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Fractal Engine Dashboard" },
    { name: "description", content: "Welcome to Fractal Engine Dashboard!" },
  ];
}
  
export default function Home() {
  return <Welcome />;
}

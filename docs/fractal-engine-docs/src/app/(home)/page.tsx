import Link from "next/link";
import Image from "next/image";
import { BookOpenText } from "lucide-react";

import FractalEngineLogo from "../../../public/static/image/fractal.svg"

export default function HomePage() {
  return (
    <div className="flex flex-col items-center justify-center flex-grow bg-gray-50">
      <div className="container mx-auto px-4 py-12">
        <Image
          src={FractalEngineLogo}
          alt="Fractal Engine Logo"
          className="mx-auto pb-4 w-[60%]"
        />

        <p className="text-2xl text-gray-500 text-center mb-12">RWAs for everyone.</p>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-5xl mx-auto">
          <Link 
            href="/docs"
            className="flex flex-col items-center p-8 bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow duration-300"
          >
            <div className="h-48 flex items-center justify-center mb-4">
              {/* <Image
                src="/static/image/fractal-docs.svg"
                alt="Fractal Engine Documentation"
                width={300}
                height={300}
                className="w-48 object-contain"
              /> */}
              <BookOpenText width={200} height={200} />
            </div>
            <h2 className="text-xl font-semibold text-center mt-auto">View Docs</h2>
          </Link>
          
          <Link 
            href="https://discord.gg/VEUMWpThg9"
            target="_blank"
            rel="noopener noreferrer"
            className="flex flex-col items-center p-8 bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow duration-300"
          >
            <div className="h-48 flex items-center justify-center mb-4">
              <Image 
                src="/static/image/Discord-Symbol-Blurple.svg" 
                alt="Discord Logo" 
                width={64} 
                height={64}
                className="w-48 h-48 object-contain"
              />
            </div>
            <h2 className="text-xl font-semibold text-center mt-auto">Join Discord</h2>
          </Link>
        </div>
      </div>
    </div>
  );
}
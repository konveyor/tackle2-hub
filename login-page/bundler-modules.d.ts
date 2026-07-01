// Declare TypeScript modules for files rspack will process but TypeScript doesn't know about.

declare module "*.svg" {
  import * as React from "react";

  export const ReactComponent: React.FC<
    React.SVGProps<SVGSVGElement> & { title?: string }
  >;

  const src: string;
  export default src;
}

declare module "*.css" {
  const classes: { readonly [key: string]: string };
  export default classes;
}

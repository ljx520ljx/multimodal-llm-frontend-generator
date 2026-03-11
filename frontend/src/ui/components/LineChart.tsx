"use client";
/*
 * Documentation:
 * Line Chart — https://app.subframe.com/f715419a5932/library?component=Line+Chart_22944dd2-3cdd-42fd-913a-1b11a3c1d16d
 */

import React from "react";
import * as SubframeCore from "@subframe/core";
import * as SubframeUtils from "../utils";

interface LineChartRootProps
  extends React.ComponentProps<typeof SubframeCore.LineChart> {
  className?: string;
}

const LineChartRoot = React.forwardRef<
  React.ElementRef<typeof SubframeCore.LineChart>,
  LineChartRootProps
>(function LineChartRoot(
  { className, ...otherProps }: LineChartRootProps,
  ref
) {
  return (
    <SubframeCore.LineChart
      className={SubframeUtils.twClassNames("h-80 w-full", className)}
      ref={ref}
      colors={[
        "#0954a5",
        "#0f3058",
        "#0091ff",
        "#0d3868",
        "#369eff",
        "#0a4481",
      ]}
      dark={true}
      {...otherProps}
    />
  );
});

export const LineChart = LineChartRoot;

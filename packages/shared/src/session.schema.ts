import { z } from "zod";

export const SessionMode = z.enum(["idle", "goal", "loop"]);

export const SessionState = z.object({
  id: z.string().min(1),
  workspaceRoot: z.string().min(1),
  mode: SessionMode.default("idle"),
  yolo: z.boolean().default(false),
  swarmEnabled: z.boolean().default(false),
  activeGoal: z.string().optional(),
  loopTask: z.string().optional(),
  loopRemaining: z.number().int().min(0).default(0),
  createdAt: z.string().min(1),
  updatedAt: z.string().min(1)
});

export type SessionState = z.infer<typeof SessionState>;

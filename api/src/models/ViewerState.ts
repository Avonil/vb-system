import mongoose, { Schema, Document } from 'mongoose';

export interface IViewerState extends Document {
  streamerId: string; // Twitch ID стрімера
  viewerId: string;   // Twitch ID глядача
  viewerName: string; // Нік глядача
  variables: Map<string, any>; // Динамічні змінні (RPG, Економіка)
  updatedAt: Date;
}

const ViewerStateSchema = new Schema<IViewerState>({
  streamerId: { type: String, required: true, index: true },
  viewerId: { type: String, required: true, index: true },
  viewerName: { type: String },
  
  // Mixed type дозволяє зберігати будь-що (числа, строки, інші об'єкти)
  variables: { type: Map, of: Schema.Types.Mixed, default: {} }
}, {
  timestamps: true // Автоматично додасть createdAt та updatedAt
});

// Унікальний індекс, щоб у стрімера не було дублікатів одного глядача
ViewerStateSchema.index({ streamerId: 1, viewerId: 1 }, { unique: true });

export const ViewerState = mongoose.model<IViewerState>('ViewerState', ViewerStateSchema);
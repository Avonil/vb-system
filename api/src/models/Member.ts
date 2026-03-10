import mongoose, { Schema, Document } from 'mongoose';

export interface IMember extends Document {
  channelId: string; // Twitch ID стрімера
  userId: string;    // Twitch ID глядача
  username: string;  // Нік глядача (кешуємо для зручності)
  
  stats: {
    xp: number;
    level: number;
    points: number;
    messagesSent: number;
  };
  
  lastSeen: Date;
}

const MemberSchema = new Schema<IMember>({
  channelId: { type: String, required: true, index: true },
  userId: { type: String, required: true, index: true },
  username: { type: String },

  stats: {
    xp: { type: Number, default: 0 },
    level: { type: Number, default: 1 },
    points: { type: Number, default: 0 },
    messagesSent: { type: Number, default: 0 }
  },

  lastSeen: { type: Date, default: Date.now }
}, {
  timestamps: true
});

// Композитний індекс для швидкого пошуку "Глядач на Каналі"
MemberSchema.index({ channelId: 1, userId: 1 }, { unique: true });

export const Member = mongoose.model<IMember>('Member', MemberSchema);
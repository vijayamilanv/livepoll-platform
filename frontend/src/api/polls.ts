import api from './axios';

export interface PollOption {
  id: string;
  text: string;
}

export interface VoteCount {
  optionId: string;
  count: number;
}

export interface Poll {
  id: string;
  shareCode: string;
  question: string;
  options: PollOption[];
  createdBy: string;
  isActive: boolean;
  createdAt: string;
  closesAt?: string;
}

export interface PollWithCounts extends Poll {
  counts: VoteCount[];
}

export const createPoll = (question: string, options: string[]) =>
  api.post<Poll>('/polls', { question, options });

export const getPoll = (shareCode: string) =>
  api.get<PollWithCounts>(`/polls/${shareCode}`);

export const getMyPolls = () =>
  api.get<{ polls: PollWithCounts[] }>('/polls/mine');

export const castVote = (shareCode: string, optionId: string) =>
  api.post<{ counts: VoteCount[] }>(`/polls/${shareCode}/vote`, { optionId });

import { createBrowserRouter, Navigate } from 'react-router-dom';
import { MainLayout } from '@/components/layout/MainLayout';
import { RepoListPage, RepoCreatePage, RepoConfigPage } from '@/pages/repos';
import { ReviewListPage } from '@/pages/reviews';
import { FeedbackListPage } from '@/pages/feedbacks';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <MainLayout />,
    children: [
      { index: true, element: <Navigate to="/repos" replace /> },
      {
        path: 'repos',
        children: [
          { index: true, element: <RepoListPage /> },
          { path: 'new', element: <RepoCreatePage /> },
          { path: ':id', element: <RepoConfigPage /> },
        ],
      },
      { path: 'reviews', element: <ReviewListPage /> },
      { path: 'feedbacks', element: <FeedbackListPage /> },
    ],
  },
]);

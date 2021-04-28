# frozen_string_literal: true

Rails.application.routes.draw do
  devise_for :users, path: 'users', controllers: {
    confirmations: 'confirmations',
    sessions: 'users/sessions',
    registrations: 'users/registrations',
    passwords: 'users/passwords'
  }

  authenticated :user do
    root to: 'portal#show', as: :authenticated_root, via: :get
  end

  match '/portal', to: 'portal#show', via: :get

  devise_scope :user do
    root to: "devise/sessions#new"
  end

  if Rails.env.development?
    require 'sidekiq/web'
    mount Sidekiq::Web, at: '/sidekiq'
  end
end
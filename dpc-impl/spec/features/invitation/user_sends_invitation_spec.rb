# frozen_string_literal: true

require 'rails_helper'

RSpec.feature 'user sends invitation to DPC' do
  include ApiClientSupport
  include MailerHelper

  let (:user) { create :user }

  context 'successful invite' do
    before(:each) do
      stub_api_client(message: :create_implementer, success: true, response: default_imp_creation_response)
  
      api_client = instance_double(ApiClient)
      allow(ApiClient).to receive(:new).and_return(api_client)
      allow(api_client).to receive(:get_provider_orgs)
        .with(user.implementer_id)
        .and_return(api_client)
      allow(api_client).to receive(:response_successful?).and_return(false)
      allow(api_client).to receive(:response_body).and_return(nil)
    end

    scenario 'user successfully sends invite' do
      sign_in user
      visit members_path

      fill_in 'user_first_name', with: 'Manny'
      fill_in 'user_last_name', with: 'York'
      fill_in 'user_email', with: 'manhattan@gmail.com'
      find('input[data-test="submit"]').click

      expect(page.body).to have_content('User invited.')
    end

    scenario 'user accepts invitation and creates an account' do
      old_user = user
      invited_user = User.invite!(first_name: 'Brook', last_name: 'York',
                                  email: 'brooklyn@gmail.com', implementer: old_user.implementer,
                                  implementer_id: old_user.implementer_id, invited_by_id: old_user.id)

      itoken = invited_user.raw_invitation_token

      visit "/impl/users/invitation/accept?invitation_token=#{itoken}"

      expect(page.body).to have_content('A member from your team has invited you to create an account in the DPC sandbox portal.')

      fill_in :user_password, with: 'Br00k1yn53^3R100!'
      fill_in :user_password_confirmation, with: 'Br00k1yn53^3R100!'
      check :user_agree_to_terms
      find('input[data-test="submit"]').click

      expect(page.body).to have_content('Your password was set successfully. You are now signed in.')
      expect(page.body).to have_content("Welcome #{invited_user.name}")

      expect(invited_user.implementer_id).to match(old_user.implementer_id)
      expect(invited_user.implementer).to match(old_user.implementer)
    end
  end

  context 'unsuccessful invite' do
    before(:each) do
      sign_in user
      visit members_path
    end

    scenario 'user does not fill out the invite form correctly' do
      find('input[data-test="submit"]').click

      expect(page.body).to have_content('All fields are required to invite a new user.')

      fill_in 'user_first_name', with: 'Manny'
      fill_in 'user_last_name', with: 'York'
      fill_in 'user_email', with: 'manny@hogwarts.edu'
      find('input[data-test="submit"]').click

      expect(page.body).to have_content('Email must be valid.')

      fill_in 'user_first_name', with: 'Manny'
      fill_in 'user_last_name', with: 'York'
      fill_in 'user_email', with: 'manhattan@gmail.com'
      find('input[data-test="submit"]').click

      expect(page.body).to have_content('User invited.')
    end

    scenario 'user invites a user already signed up' do
      user1 = create(:user, email: 'email@gmail.com')

      fill_in 'user_first_name', with: user1.first_name
      fill_in 'user_last_name', with: user1.last_name
      fill_in 'user_email', with: user1.email
      find('input[data-test="submit"]').click

      expect(page.body).to have_content('User already has an account.')
    end
  end

  context 'successfully resend invite' do
    scenario 'invited user requests new invite' do
      user = create(:user, invitation_sent_at: DateTime.now, invitation_accepted_at: nil)

      visit new_user_confirmation_path

      fill_in 'user_email', with: user.email
      find('input[data-test="submit"]').click

      last_delivery = ActionMailer::Base.deliveries.last

      expect(last_delivery).not_to be_nil
      expect(last_delivery.to).to include(user.email)
      expect(last_delivery.subject).to include('Invitation instructions')
    end
  end
end
